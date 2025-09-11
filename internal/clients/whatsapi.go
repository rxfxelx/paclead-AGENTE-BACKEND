package clients

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type Whats struct {
    Base  string
    Token string
    http  *http.Client
}

func NewWhats(base, token string) *Whats {
    return &Whats{Base: trimSlash(base), Token: token, http: &http.Client{}}
}

// trimSlash remove barras finais de uma URL base para evitar // duplicadas.
func trimSlash(s string) string {
    for len(s) > 0 && s[len(s)-1] == '/' {
        s = s[:len(s)-1]
    }
    return s
}

func (w *Whats) do(ctx context.Context, path string, body any) error {
    buf, _ := json.Marshal(body)
    req, _ := http.NewRequestWithContext(ctx, "POST", w.Base+path, bytes.NewReader(buf))
    // Alguns provedores usam header "token"; outros "Authorization: Bearer"
    req.Header.Set("token", w.Token)
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")
    resp, err := w.http.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= http.StatusMultipleChoices {
        return fmt.Errorf("whats api %s: status %d", path, resp.StatusCode)
    }
    return nil
}

func (w *Whats) SendText(ctx context.Context, number, text string) error {
    return w.do(ctx, "/send/text", map[string]any{
        "number": number,
        "text":   text,
    })
}

func (w *Whats) SendCarousel(ctx context.Context, number, text string, cards []map[string]any) error {
    return w.do(ctx, "/send/carousel", map[string]any{
        "number":  number,
        "text":    text,
        "carousel": cards,
        "delay":   0,
        "readchat": true,
    })
}

func (w *Whats) SendAudioBase64(ctx context.Context, number, b64 string) error {
    // Example for Zapster-like endpoint: adapt as needed.
    // If your Whats gateway accepts audio via the same base URL, adjust path here.
    return w.do(ctx, "/send/audio", map[string]any{
        "number": number,
        "file":   b64,
        "type":   "audio",
    })
}
