package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"pac-lead-agent/internal/types"
)

type PacLead struct {
	Base      string // :8889 (produtos/leads)
	CRM       string // :8082 (CRM)
	Platform  string // plataforma backend (ex: https://.../api)
	http      *http.Client
}

func NewPacLead(base, crm, platform string) *PacLead {
	return &PacLead{
		Base:     trim(base),
		CRM:      trim(crm),
		Platform: trim(platform),
		http:     &http.Client{},
	}
}

// trim removes trailing slashes from a base URL.
func trim(s string) string {
	return strings.TrimRight(s, "/")
}

func (p *PacLead) postJSON(ctx context.Context, url string, in any, out any) error {
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(in)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, &buf)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	io.Copy(io.Discard, resp.Body)
	return nil
}

func (p *PacLead) LeadsGeral(ctx context.Context, numero, cnpj string) (map[string]any, error) {
	url := p.Base + "/leads_geral"
	in := map[string]any{
		"id": 0, "nome": "", "numero": numero, "status": 0, "lead": 0,
		"Thread_id": "", "data_ult_msg": "", "ult_msg_numero": "", "cnpj_cpf": cnpj,
	}
	var out map[string]any
	err := p.postJSON(ctx, url, in, &out)
	return out, err
}

func (p *PacLead) LeadPost(ctx context.Context, lead types.LeadRecord) (map[string]any, error) {
	url := p.Base + "/leadpost"
	var out map[string]any
	err := p.postJSON(ctx, url, lead, &out)
	return out, err
}

func (p *PacLead) Produtos(ctx context.Context, cnpj string, id *string) ([]map[string]any, error) {
	url := fmt.Sprintf("%s/produtos?cnpj=%s", p.Base, cnpj)
	if id != nil && *id != "" {
		url += "&id=" + *id
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateCRMLead mirrors the :8082 /leads update payload used in the workflow.
// Provide exactly the fields your CRM expects.
func (p *PacLead) UpdateCRMLead(ctx context.Context, payload any) error {
	// Se CRM não configurado, não faz nada
	if p.CRM == "" {
		return nil
	}
	url := p.CRM + "/leads"
	return p.postJSON(ctx, url, payload, nil)
}

// GetAgentSettings consulta a plataforma por configurações do agente (prompt custom por tenant).
// Espera endpoint: GET {Platform}/api/agent/settings?org_id=...&flow_id=...
// Retorna o objeto JSON (ou nil/falha silenciosa se não configurado).
func (p *PacLead) GetAgentSettings(ctx context.Context, orgID, flowID string) (map[string]any, error) {
	if strings.TrimSpace(p.Platform) == "" {
		return nil, nil
	}
	q := url.Values{}
	if strings.TrimSpace(orgID) != "" {
		q.Set("org_id", orgID)
	}
	if strings.TrimSpace(flowID) != "" {
		q.Set("flow_id", flowID)
	}
	u := p.Platform + "/api/agent/settings"
	if s := q.Encode(); s != "" {
		u += "?" + s
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	if orgID != "" {
		req.Header.Set("X-Org-ID", orgID)
	}
	if flowID != "" {
		req.Header.Set("X-Flow-ID", flowID)
	}

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("settings http %d", resp.StatusCode)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	// Alguns backends respondem { data: {...} }
	if data, ok := out["data"].(map[string]any); ok {
		return data, nil
	}
	return out, nil
}
