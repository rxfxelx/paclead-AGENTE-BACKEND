package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Platform struct {
	Base    string
	client  *http.Client
}

func NewPlatform(base string) *Platform {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "http://localhost:8080"
	}
	return &Platform{
		Base:   strings.TrimRight(base, "/"),
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// GetAgentSettings busca /api/agent/settings na Plataforma usando os headers X-Org-ID e X-Flow-ID.
func (p *Platform) GetAgentSettings(ctx context.Context, orgID, flowID string) (map[string]any, error) {
	if p == nil || p.Base == "" {
		return nil, fmt.Errorf("platform base url not configured")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", p.Base+"/api/agent/settings", nil)
	if err != nil {
		return nil, err
    }
	// Multi-tenant headers
	if strings.TrimSpace(orgID) != "" {
		req.Header.Set("X-Org-ID", orgID)
	}
	if strings.TrimSpace(flowID) != "" {
		req.Header.Set("X-Flow-ID", flowID)
	}
	req.Header.Set("Accept", "application/json")

	res, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Se a plataforma ainda não tem registro, retornamos nil para o caller usar fallback.
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("platform settings http %d", res.StatusCode)
	}

	var out map[string]any
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
