package clients

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"

    "pac-lead-agent/internal/types"
)

type PacLead struct {
    Base string // :8889
    CRM  string // :8082
    http *http.Client
}

func NewPacLead(base, crm string) *PacLead {
    return &PacLead{Base: trim(base), CRM: trim(crm), http: &http.Client{}}
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
