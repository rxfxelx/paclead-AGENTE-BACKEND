package httpapi

import (
    "encoding/json"
    "log"
    "net/http"

    "pac-lead-agent/internal/config"
    "pac-lead-agent/internal/flow"
    "pac-lead-agent/internal/types"
)

func RegisterRoutes(mux *http.ServeMux, cfg config.Config) {
    h := &handler{cfg: cfg}
    mux.HandleFunc("/webhooks/paclead-maryjoias", h.webhook)
}

type handler struct {
    cfg config.Config
}

func (h *handler) webhook(w http.ResponseWriter, r *http.Request) {
    var payload types.IncomingWebhook
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "bad payload", http.StatusBadRequest)
        return
    }

    resp, err := flow.HandleIncomingMessage(r.Context(), h.cfg, payload)
    if err != nil {
        log.Println("flow error:", err)
    }

    _ = json.NewEncoder(w).Encode(resp)
}
