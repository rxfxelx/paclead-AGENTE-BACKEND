package flow

// Options carrega contexto da instância (id / token) para a resposta via WhatsApp.
type Options struct {
	InstanceID    string
	InstanceToken string
}

type Option func(*Options)

func WithInstance(id, token string) Option {
	return func(o *Options) {
		o.InstanceID = id
		o.InstanceToken = token
	}
}
