package config

import "os"

type Config struct {
	Addr               string
	OpenAIKey          string
	OpenAIAssistantID  string
	UAzapiToken        string
	UAzapiBaseURL      string
	PacLeadBaseURL     string
	PacLeadCRMBaseURL  string
	PlatformBaseURL    string // <— NOVO: backend da plataforma (para /api/agent/settings)
	DefaultPrompt      string // <— NOVO: fallback de prompt padrão
	RedisURL           string
}

func Load() Config {
	return Config{
		Addr:              getenv("APP_ADDR", ":8080"),
		OpenAIKey:         os.Getenv("OPENAI_API_KEY"),
		OpenAIAssistantID: getenv("OPENAI_ASSISTANT_ID", "asst_xxx"),
		UAzapiToken:       os.Getenv("UAZAPI_TOKEN"),
		UAzapiBaseURL:     getenv("UAZAPI_BASE_URL", "https://hia-clientes.uazapi.com"),
		PacLeadBaseURL:    getenv("PACLEAD_BASE_URL", "http://paclead.com.br:8889"),
		PacLeadCRMBaseURL: getenv("PACLEAD_CRM_BASE_URL", "https://paclead.com.br:8082"),
		PlatformBaseURL:   getenv("PLATFORM_BASE_URL", ""),     // e.g. https://plataforma-pac-lead-backend-production.up.railway.app
		DefaultPrompt:     getenv("DEFAULT_PROMPT", ""),        // se vazio, usamos o default embarcado
		RedisURL:          os.Getenv("REDIS_URL"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
