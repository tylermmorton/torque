package torque

type Mode string

const (
	ModeDevelopment Mode = "development"
	ModeProduction  Mode = "production"
)

//type Option func(h *controllerImpl[T])
//
//func WithMode(mode Mode) Option {
//	return func(h *controllerImpl[T]) {
//		h.mode = mode
//	}
//}
