package model

const (
	CORS           Middleware = 1
	NoCaching      Middleware = 2
	Counter        Middleware = 3
	Histogram      Middleware = 4
	PanicTo500     Middleware = 5
	RequestLogging Middleware = 6
)

type (
	Middleware int

	MiddlewareWrapper interface {
		Wrap(subsystem, name string, middleware Middleware, handler Handle) Handle
	}
)
