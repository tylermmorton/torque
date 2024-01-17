package torque

type Mode string

const (
	ModeDevelopment Mode = "development"
	ModeProduction  Mode = "production"
)

type Option func(h *moduleHandler)

func WithMode(mode Mode) Option {
	return func(h *moduleHandler) {
		h.mode = mode
	}
}
