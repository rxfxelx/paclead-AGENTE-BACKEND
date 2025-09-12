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

func HandleIncomingMessage(ctx context.Context, cfg config.Config, in types.IncomingWebhook, opts ...Option) (Response, error) {
	// aplica opções (instância, tenant, slug)
	var o Options
	for _, fn := range opts {
		fn(&o)
	}

	// escolhe o token da instância se vier do header; caso contrário, usa o global do cfg
	token := cfg.UAzapiToken
	if strings.TrimSpace(o.InstanceToken) != "" {
		token = o.InstanceToken
	}

	whats := clients.NewWhats(cfg.UAzapiBaseURL, token)
	ai := clients.NewOpenAI(cfg.OpenAIKey, cfg.OpenAIAssistantID)
	pl := clients.NewPacLead(cfg.PacLeadBaseURL, cfg.PacLeadCRMBaseURL, cfg.PlatformBaseURL)

	number := extractNumber(in.Body.Message.ChatID)
	text := strings.TrimSpace(in.Body.Message.Content)
	msgType := strings.ToLower(in.Body.Message.Type)

	// Ajusta CNPJ (multi-tenant) via settings; fallback mantém constante
	cnpj := "23820015000100"
	if settings, err := pl.GetAgentSettings(ctx, o.OrgID, o.FlowID); err == nil && settings != nil {
		if v, ok := settings["tax_id"].(string); ok && strings.TrimSpace(v) != "" {
			cnpj = onlyDigits(v)
		}
	}

	// Obtém prompt final (DEFAULT + customizações do cliente)
	prompt, _ := BuildPrompt(ctx, cfg, pl, o.OrgID, o.FlowID)

	threadID, err := EnsureThread(ctx, ai, pl, number, cnpj)
	if err != nil {
		return Response{}, err
	}

	switch msgType {
	case "text", "conversation", "extendedtextmessage", "templatebuttonreplymessage":
		if text != "" {
			// Se mensagem do usuário já veio com "ID_P:" envia carrossel de produtos direto
			if ids := parseIDs(strings.ToUpper(text)); len(ids) > 0 {
				_ = whats.SendText(ctx, number, "Procurando produtos…")
				_ = SendProductsCarousel(ctx, pl, whats, cnpj, number, ids)
				return Response{Ok: true}, nil
			}

			// Envia mensagem do usuário e cria run com 'instructions' = prompt final
			if err := SendUserTextAndRunWithInstructions(ctx, ai, threadID, text, prompt); err == nil {
				if reply, _ := GetLastAssistantText(ctx, ai, threadID); reply != "" {
					// Se a resposta do assistente contiver "ID_P:", enviamos o carrossel
					if ids := parseIDs(strings.ToUpper(reply)); len(ids) > 0 {
						_ = whats.SendText(ctx, number, "Separei alguns produtos para você 👇")
						_ = SendProductsCarousel(ctx, pl, whats, cnpj, number, ids)
					} else {
						_ = whats.SendText(ctx, number, reply)
					}
				}
			}
		}
	case "image":
		// Ponto de entrada para visão — por enquanto responde texto
		_ = whats.SendText(ctx, number, "📸 Recebi a imagem! Vou analisar e já retorno.")
	case "audio", "audiomessage", "ptt":
		// Responde com áudio da última mensagem do assistente
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
	if i == -1 {
		return nil
	}
	key := strings.TrimSpace(s[:i])
	if key != "ID_P" {
		return nil
	}
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

// onlyDigits remove qualquer caractere não numérico (útil para CPF/CNPJ)
func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
