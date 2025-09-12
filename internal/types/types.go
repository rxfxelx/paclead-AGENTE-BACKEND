package types

import (
	"encoding/json"
)

// IncomingWebhook representa um payload minimamente compatível com Uazapi,
// conforme usado pelo dispatcher (chatId, type, content).
type IncomingWebhook struct {
	Instance string      `json:"instance,omitempty"`
	Event    string      `json:"event,omitempty"`
	Body     IncomingBody`json:"body"`
}

type IncomingBody struct {
	Message Message `json:"message"`
	// Outros campos podem existir; ignorados por ora
}

type Message struct {
	ChatID  string `json:"chatId"`
	Type    string `json:"type"`
	Content string `json:"content"`
	// Campos adicionais ignorados
}

// Unmarshal robusto para aceitar variações (chat_id, remoteJid, etc.)
func (m *Message) UnmarshalJSON(data []byte) error {
	type alias Message
	var a alias
	if err := json.Unmarshal(data, &a); err == nil && (a.ChatID != "" || a.Type != "" || a.Content != "") {
		*m = Message(a)
		// preenchido diretamente
	}

	// Fallback: map genérico para tentar chaves alternativas
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil // melhor não falhar parsing
	}
	str := func(v any) string {
		if v == nil {
			return ""
		}
		if s, ok := v.(string); ok {
			return s
		}
		return ""
	}

	if m.ChatID == "" {
		m.ChatID = str(raw["chatId"])
	}
	if m.ChatID == "" {
		m.ChatID = str(raw["chat_id"])
	}
	if m.ChatID == "" {
		m.ChatID = str(raw["remoteJid"])
	}
	if m.Type == "" {
		m.Type = str(raw["type"])
	}
	if m.Content == "" {
		m.Content = str(raw["content"])
	}
	return nil
}

// LeadRecord espelha o contrato usado pela API /leadpost do serviço PACLEAD (:8889)
type LeadRecord struct {
	ID         int    `json:"id"`
	Nome       string `json:"nome"`
	Numero     string `json:"numero"`
	Status     int    `json:"status"`
	Lead       int    `json:"lead"`
	ThreadID   string `json:"Thread_id"`
	DataUltMsg string `json:"data_ult_msg"`
	UltMsgNum  string `json:"ult_msg_numero"`
	CNPJCPF    string `json:"cnpj_cpf"`
}
