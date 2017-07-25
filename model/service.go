package model

type (
	ShutdownFunc func(log Logger)

	ServiceOptions struct {
		Logger                Logger
		Metrics               Metrics
		RouterFactory         RouterFactory
		ServiceHandlerFactory ServiceHandlerFactory
		ShutdownFunc          ShutdownFunc
	}

	Service interface {
		Run()
		AddRoute(name string, routes []string, methods []string, middlewares []Middleware, handler Handle)
	}
)
