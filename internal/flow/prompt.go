package flow

import "strings"

// MergePrompt é um utilitário opcional para compor prompts adicionais de forma segura.
// Não conflita com BuildPrompt (que já está no options.go).
func MergePrompt(base string, extras ...string) string {
	base = strings.TrimSpace(base)
	var b strings.Builder
	b.WriteString(base)
	for _, e := range extras {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(e)
	}
	return b.String()
}
