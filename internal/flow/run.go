package flow

import (
    "context"
    "time"

    "pac-lead-agent/internal/clients"
    "pac-lead-agent/internal/types"
)

func EnsureThread(ctx context.Context, ai *clients.OpenAI, pl *clients.PacLead, number, cnpj string) (string, error) {
    _, _ = pl.LeadsGeral(ctx, number, cnpj)
    tid, err := ai.CreateThread(ctx)
    if err != nil { return "", err }
    _, _ = pl.LeadPost(ctx, types.LeadRecord{
        ID: 0, Nome: "", Numero: number, Status: 1, Lead: 1,
        ThreadID: tid, DataUltMsg: time.Now().Format("2006-01-02 15:04"), UltMsgNum: "", CNPJCPF: cnpj,
    })
    return tid, nil
}

func SendUserTextAndRun(ctx context.Context, ai *clients.OpenAI, threadID, text string) error {
    content := []map[string]any{{"type": "text", "text": text}}
    if err := ai.CreateMessage(ctx, threadID, "user", content); err != nil {
        return err
    }
    _, err := ai.CreateRun(ctx, threadID)
    return err
}

func GetLastAssistantText(ctx context.Context, ai *clients.OpenAI, threadID string) (string, error) {
    return ai.LastMessageText(ctx, threadID)
}
