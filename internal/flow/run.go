package flow

import (
	"context"
	"time"

	"pac-lead-agent/internal/clients"
	"pac-lead-agent/internal/types"
)

func EnsureThread(ctx context.Context, ai *clients.OpenAI, pl *clients.PacLead, number, cnpj string) (string, error) {
	// Tenta recuperar lead existente e reaproveitar Thread_id
	if out, err := pl.LeadsGeral(ctx, number, cnpj); err == nil && out != nil {
		for _, k := range []string{"Thread_id", "thread_id", "thread", "ThreadID"} {
			if v, ok := out[k]; ok {
				if s, ok := v.(string); ok && s != "" {
					return s, nil
				}
			}
		}
	}
	// Cria nova thread e salva no lead
	tid, err := ai.CreateThread(ctx)
	if err != nil {
		return "", err
	}
	_, _ = pl.LeadPost(ctx, types.LeadRecord{
		ID:         0,
		Nome:       "",
		Numero:     number,
		Status:     1,
		Lead:       1,
		ThreadID:   tid,
		DataUltMsg: time.Now().Format("2006-01-02 15:04"),
		UltMsgNum:  "",
		CNPJCPF:    cnpj,
	})
	return tid, nil
}

// Mantém a função original (compatibilidade)
func SendUserTextAndRun(ctx context.Context, ai *clients.OpenAI, threadID, text string) error {
	content := []map[string]any{{"type": "text", "text": text}}
	if err := ai.CreateMessage(ctx, threadID, "user", content); err != nil {
		return err
	}
	_, err := ai.CreateRun(ctx, threadID)
	return err
}

// Nova função: cria Run com 'instructions' = prompt final (override dinâmico)
func SendUserTextAndRunWithInstructions(ctx context.Context, ai *clients.OpenAI, threadID, text, instructions string) error {
	content := []map[string]any{{"type": "text", "text": text}}
	if err := ai.CreateMessage(ctx, threadID, "user", content); err != nil {
		return err
	}
	_, err := ai.CreateRunWithInstructions(ctx, threadID, instructions)
	return err
}

func GetLastAssistantText(ctx context.Context, ai *clients.OpenAI, threadID string) (string, error) {
	return ai.LastMessageText(ctx, threadID)
}
