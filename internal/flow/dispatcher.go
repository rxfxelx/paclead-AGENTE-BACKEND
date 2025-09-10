package flow

import (
    "context"
    "fmt"
    "strings"

    "pac-lead-agent/internal/clients"
    "pac-lead-agent/internal/config"
    "pac-lead-agent/internal/types"
)

type Response struct {
    Ok bool `json:"ok"`
}

func HandleIncomingMessage(ctx context.Context, cfg config.Config, in types.IncomingWebhook) (Response, error) {
    whats := clients.NewWhats(cfg.UAzapiBaseURL, cfg.UAzapiToken)
    ai := clients.NewOpenAI(cfg.OpenAIKey, cfg.OpenAIAssistantID)
    pl := clients.NewPacLead(cfg.PacLeadBaseURL, cfg.PacLeadCRMBaseURL)

    number := extractNumber(in.Body.Message.ChatID)
    text := strings.TrimSpace(in.Body.Message.Content)
    msgType := strings.ToLower(in.Body.Message.Type)

    threadID, err := EnsureThread(ctx, ai, pl, number, "23820015000100")
    if err != nil { return Response{}, err }

    switch msgType {
    case "text", "conversation", "extendedtextmessage", "templatebuttonreplymessage":
        if text != "" {
            if err := SendUserTextAndRun(ctx, ai, threadID, text); err == nil {
                if reply, _ := GetLastAssistantText(ctx, ai, threadID); reply != "" {
                    _ = whats.SendText(ctx, number, reply)
                }
            }
        }
    case "image":
        // TODO: vision support if needed
        _ = whats.SendText(ctx, number, "Recebi a imagem! Em breve descrevo pra você.")
    case "audio", "audiomessage":
        _ = SendAssistantReplyAudio(ctx, ai, whats, threadID, number)
    default:
        _ = whats.SendText(ctx, number, fmt.Sprintf("Tipo de mensagem não suportado ainda: %s", msgType))
    }
    return Response{Ok: true}, nil
}

func extractNumber(chatid string) string {
    if i := strings.IndexByte(chatid, '@'); i > 0 {
        return chatid[:i]
    }
    return chatid
}

// Helper: parse ids from a string like "ID_P: 1, 2, 3"
func parseIDs(s string) []string {
    s = strings.TrimSpace(s)
    i := strings.Index(s, ":")
    if i == -1 { return nil }
    rest := s[i+1:]
    parts := strings.Split(rest, ",")
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        if x := strings.TrimSpace(p); x != "" {
            out = append(out, x)
        }
    }
    return out
}
