package flow

import (
	"context"
	"fmt"
	"strings"

	"pac-lead-agent/internal/clients"
	"pac-lead-agent/internal/config"
)

// ===== Options pattern (instância / tenant / slug) =====

type Option func(*Options)

type Options struct {
	InstanceID    string
	InstanceToken string
	OrgID         string
	FlowID        string
	Slug          string
}

func WithInstance(id, token string) Option {
	return func(o *Options) {
		o.InstanceID = strings.TrimSpace(id)
		o.InstanceToken = strings.TrimSpace(token)
	}
}

func WithTenant(orgID, flowID string) Option {
	return func(o *Options) {
		o.OrgID = strings.TrimSpace(orgID)
		o.FlowID = strings.TrimSpace(flowID)
	}
}

func WithSlug(slug string) Option {
	return func(o *Options) {
		o.Slug = strings.TrimSpace(slug)
	}
}

// ===== Prompt Builder =====

// BuildPrompt compõe o prompt final a partir de um DEFAULT_PROMPT + customizações do cliente
// obtidas via Platform (/api/agent/settings). É tolerante a falhas: se não houver plataforma
// configurada ou o endpoint indisponível, retorna apenas o prompt padrão.
func BuildPrompt(ctx context.Context, cfg config.Config, pl *clients.PacLead, orgID, flowID string) (string, error) {
	base := strings.TrimSpace(cfg.DefaultPrompt)
	if base == "" {
		base = strings.TrimSpace(defaultPromptPTBR)
	}

	// Sem plataforma configurada: retorna o base.
	if pl == nil || strings.TrimSpace(pl.Platform) == "" {
		return base, nil
	}

	settings, err := pl.GetAgentSettings(ctx, orgID, flowID)
	if err != nil || settings == nil {
		return base, nil
	}

	// Extrai campos comuns (opcionais)
	get := func(k string) string {
		if v, ok := settings[k]; ok && v != nil {
			if s, ok := v.(string); ok {
				return strings.TrimSpace(s)
			}
		}
		return ""
	}
	name := get("name")
	sector := get("sector")
	style := get("communicationStyle")
	profileType := get("profileType")
	custom := get("profileCustom")
	baseOverlay := get("basePrompt") // se o cliente quiser sobrepor grande parte do prompt

	var sb strings.Builder
	if baseOverlay != "" {
		sb.WriteString(strings.TrimSpace(baseOverlay))
	} else {
		sb.WriteString(base)
	}

	sb.WriteString("\n\n")
	sb.WriteString("### Contexto do cliente (multi-tenant)\n")
	if name != "" {
		sb.WriteString(fmt.Sprintf("- Nome do agente: %s\n", name))
	}
	if sector != "" {
		sb.WriteString(fmt.Sprintf("- Setor/Indústria: %s\n", sector))
	}
	if style != "" {
		sb.WriteString(fmt.Sprintf("- Estilo de comunicação preferido: %s\n", style))
	}
	if profileType != "" {
		sb.WriteString(fmt.Sprintf("- Tipo de perfil: %s\n", profileType))
	}
	if custom != "" {
		sb.WriteString(fmt.Sprintf("- Instruções adicionais do cliente: %s\n", custom))
	}

	return sb.String(), nil
}

// Prompt padrão (fallback) — claro, direto e com protocolo de produtos.
// O Assistente pode recomendar, qualificar e vender. Para mostrar produtos
// no WhatsApp, ELE DEVE imprimir uma linha 'ID_P: 10, 22' com os IDs
// dos produtos. O backend detecta essa linha e envia automaticamente
// um carrossel nativo no WhatsApp.
const defaultPromptPTBR = `
Você é **Pac Lead**, um agente de vendas especializado em atendimento comercial no WhatsApp.
Fale SEMPRE em português do Brasil, com naturalidade, clareza e objetividade.
Seu objetivo é QUALIFICAR, RECOMENDAR e VENDER, com foco no que o cliente precisa.

**Regras de ouro**
1. Seja proativo e amigável; abra a conversa, faça perguntas abertas e avance para a venda.
2. Se o cliente pedir produtos ou preços, ofereça os itens mais relevantes.
3. Para exibir produtos no WhatsApp, escreva **UMA LINHA** com o padrão:
   ID_P: <id1>, <id2>, <id3>
   - Ex.: "ID_P: 12, 33, 71"
   - Essa linha NÃO deve ter texto extra; o sistema detecta e envia um carrossel com imagens e preços.
4. Depois do carrossel, convide o cliente a fechar a compra (“Posso emitir agora?”, “Qual forma de pagamento?”).
5. Se o cliente pedir algo específico (ex.: cor, tamanho), ajuste a recomendação e repasse novos IDs.
6. Mantenha postura ética e cordial; não invente informações que você não tem.

**Estratégia**
- Investigue rapidamente a necessidade (uso, orçamento, urgência).
- Explique benefícios em bullets breves.
- Se necessário, sugira variações (básico / intermediário / premium).
- Feche com CTA claro (emitir pedido, reservar, marcar retirada, etc.).

**Importante**
- Se você decidir que é hora de mostrar produtos, imprima apenas a linha com “ID_P: ...” (sem mais nada nessa linha).
- O restante da conversa continua normalmente nas mensagens seguintes.
`
