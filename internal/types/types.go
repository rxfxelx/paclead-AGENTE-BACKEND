package types

type IncomingWebhook struct {
	Body struct {
		Message struct {
			Type           string `json:"type"`
			MessageType    string `json:"messageType"`
			Content        string `json:"content"`
			MessageID      string `json:"messageid"`
			ChatID         string `json:"chatid"`
			ButtonOrListID string `json:"buttonOrListid,omitempty"`
		} `json:"message"`
		Data struct {
			Sender struct {
				ID string `json:"id"`
			} `json:"sender"`
			Content struct {
				Media struct {
					URL string `json:"url"`
				} `json:"media"`
			} `json:"content"`
		} `json:"data"`
	} `json:"body"`
}

type LeadRecord struct {
	ID         int64  `json:"id"`
	Nome       string `json:"nome"`
	Numero     string `json:"numero"`
	Status     int    `json:"status"`
	Lead       int    `json:"lead"`
	ThreadID   string `json:"Thread_id"`
	DataUltMsg string `json:"data_ult_msg"`
	UltMsgNum  string `json:"ult_msg_numero"`
	CNPJCPF    string `json:"cnpj_cpf"`
}
