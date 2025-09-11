package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"
    "syscall"

    "pac-lead-agent/internal/config"
    "pac-lead-agent/internal/httpapi"
)

func main() {
    cfg := config.Load()

    mux := http.NewServeMux()
    httpapi.RegisterRoutes(mux, cfg)

    // simple healthcheck
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })

    srv := &http.Server{
        Addr:              cfg.Addr,
        Handler:           mux,
        ReadTimeout:       30 * time.Second,
        ReadHeaderTimeout: 10 * time.Second,
        WriteTimeout:      30 * time.Second,
    }

    // run server with graceful shutdown
    errCh := make(chan error, 1)
    go func() {
        log.Printf("Pac Lead Agent listening on %s", cfg.Addr)
        errCh <- srv.ListenAndServe()
    }()
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    select {
    case <-quit:
        log.Println("shutting down...")
        _ = srv.Close()
    case err := <-errCh:
        if err != nil {
            log.Println("server error:", err)
            os.Exit(1)
        }
    }
}
