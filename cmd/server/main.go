package main

import (
    "log"
    "net/http"
    "os"
    "time"

    "pac-lead-agent/internal/config"
    "pac-lead-agent/internal/httpapi"
)

func main() {
    cfg := config.Load()

    mux := http.NewServeMux()
    httpapi.RegisterRoutes(mux, cfg)

    srv := &http.Server{
        Addr:              cfg.Addr,
        Handler:           mux,
        ReadTimeout:       30 * time.Second,
        ReadHeaderTimeout: 10 * time.Second,
        WriteTimeout:      30 * time.Second,
    }

    log.Printf("Pac Lead Agent listening on %s", cfg.Addr)
    if err := srv.ListenAndServe(); err != nil {
        log.Println("server error:", err)
        os.Exit(1)
    }
}
