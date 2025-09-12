package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Platform struct {
	baseURL string
	httpc   *http.Client
}

// NewPlatform cria um cliente para a Plataforma (ex.: https://plataforma-pac-lead-backend-production.up.railway.app)
func NewPlatform(baseURL string) *Platform {
	return &Platform{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpc:   &http.Client{Timeout: 12 * time.Second},
	}
}

func (p *Platform) endpoint(path string) string {
	if strings.HasPrefix(path, "/") {
		return p.baseURL + path
	}
	return p.baseURL + "/" + path
}

// GetAgentSettings tenta buscar as configurações do agente:
// 1) GET /api/agent-config  (se existir)
// 2) fallback: GET /api/company e normaliza atributos úteis (ex.: tax_id)
func (p *Platform) GetAgentSettings(ctx context.Context, orgID, flowID string) (map[string]any, error) {
	if p == nil || p.baseURL == "" {
		return nil, errors.New("platform baseURL not configured")
	}

	// Tenta endpoint moderno
	if settings, err := p.getJSON(ctx, p.endpoint("/api/agent-config"), orgID, flowID); err == nil && len(settings) > 0 {
		return settings, nil
	}

	// Fallback para /api/company
	if company, err := p.getJSON(ctx, p.endpoint("/api/company"), orgID, flowID); err == nil && len(company) > 0 {
		out := map[string]any{}
		if v, ok := company["tax_id"].(string); ok && strings.TrimSpace(v) != "" {
			out["tax_id"] = v
		}
		if v, ok := company["nome_fantasia"].(string); ok {
			out["display_name"] = v
		}
		if len(out) > 0 {
			return out, nil
		}
	}

	return nil, fmt.Errorf("no settings available")
}

// getJSON executa GET com headers multi-tenant e decodifica em map[string]any
func (p *Platform) getJSON(ctx context.Context, url string, orgID, flowID string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if orgID != "" {
		req.Header.Set("X-Org-ID", orgID)
	}
	if flowID != "" {
		req.Header.Set("X-Flow-ID", flowID)
	}
	req.Header.Set("Accept", "application/json")

	res, err := p.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// 404 é tratado como "não existe esse endpoint", e o caller tenta fallback
	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("404")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("platform %s: %s", res.Status, strings.TrimSpace(string(b)))
	}

	var m map[string]any
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}
