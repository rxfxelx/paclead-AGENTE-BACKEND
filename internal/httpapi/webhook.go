package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"pac-lead-agent/internal/config"
	"pac-lead-agent/internal/flow"
	"pac-lead-agent/internal/types"
)

func RegisterRoutes(mux *http.ServeMux, cfg config.Config) {
	h := &handler{cfg: cfg}
	// Compatibilidade com fluxo antigo (prefixo fixo)
	mux.HandleFunc("/webhooks/paclead-maryjoias", h.webhook)
	// Webhook padrão para Uazapi (aceita eventos diretos sem slug)
	mux.HandleFunc("/webhook/uazapi", h.webhook)
	// Webhook dinâmico: aceita /webhooks/<slug> e repassa ao handler
	mux.HandleFunc("/webhooks/", h.webhookDynamic)
}

type handler struct {
	cfg config.Config
}

func (h *handler) webhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var payload types.IncomingWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}

	instID := strings.TrimSpace(r.Header.Get("X-Instance-ID"))
	instToken := strings.TrimSpace(r.Header.Get("X-Instance-Token"))
	orgID := strings.TrimSpace(r.Header.Get("X-Org-ID"))
	flowID := strings.TrimSpace(r.Header.Get("X-Flow-ID"))

	resp, err := flow.HandleIncomingMessage(
		r.Context(),
		h.cfg,
		payload,
		flow.WithInstance(instID, instToken),
		flow.WithTenant(orgID, flowID),
	)
	if err != nil {
		log.Println("flow error:", err)
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// webhookDynamic trata caminhos /webhooks/<slug>.
// O slug não é utilizado nesta versão, mas pode ser usado para multi-tenant no futuro.
func (h *handler) webhookDynamic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Obtém o slug removendo o prefixo '/webhooks/'
	slug := strings.TrimPrefix(r.URL.Path, "/webhooks/")
	if slug == "" || slug == r.URL.Path {
		http.NotFound(w, r)
		return
	}

	var payload types.IncomingWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}

	instID := strings.TrimSpace(r.Header.Get("X-Instance-ID"))
	instToken := strings.TrimSpace(r.Header.Get("X-Instance-Token"))
	orgID := strings.TrimSpace(r.Header.Get("X-Org-ID"))
	flowID := strings.TrimSpace(r.Header.Get("X-Flow-ID"))

	resp, err := flow.HandleIncomingMessage(
		r.Context(),
		h.cfg,
		payload,
		flow.WithInstance(instID, instToken),
		flow.WithTenant(orgID, flowID),
		flow.WithSlug(slug),
	)
	if err != nil {
		log.Println("flow error:", err, "slug:", slug)
	}
	_ = json.NewEncoder(w).Encode(resp)
}
