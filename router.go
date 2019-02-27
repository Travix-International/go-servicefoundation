package servicefoundation

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type (
	// Handle is a function signature for the ServiceFoundation handlers
	Handle func(WrappedResponseWriter, *http.Request, HandlerUtils)

	// MetaFunc is a function that returns a map containing meta data used to enrich log messages.
	MetaFunc func(*http.Request, RouteParamsFunc) Meta

	// RouteParamsFunc is a function that returns a map containing the route parameters
	RouteParamsFunc func() RouteParams

	// HandlerUtils contains utilities that can be used by the current handler.
	HandlerUtils struct {
		MetaFunc   MetaFunc
		ParamsFunc RouteParamsFunc
		LogFactory LogFactory
		Metrics    Metrics
	}

	// Meta is a map containing meta data used to enrich log messages.
	Meta map[string]string

	// RouteParams is a map containing the route parameters.
	RouteParams map[string]string

	// RouterFactory is an interface to create a new Router.
	RouterFactory interface {
		NewRouter() Router
	}

	routerFactoryImpl struct {
		logFactory LogFactory
		metrics    Metrics
	}

	// Router is an interface containing the used methods from httprouter.Router
	Router interface {
		Handle(method, path string, metaFunc MetaFunc, handle Handle)
		ServeHTTP(http.ResponseWriter, *http.Request)
	}

	httpRouterImpl struct {
		logFactory LogFactory
		metrics    Metrics
		router     *httprouter.Router
	}
)

var (
	// MethodsForGet contains a slice with the supported http methods for GET.
	MethodsForGet = []string{http.MethodGet}
	// MethodsForPost contains a slice with the supported http methods for POST.
	MethodsForPost = []string{http.MethodPost}
)

// NewRouterFactory instantiates a new RouterFactory implementation.
func NewRouterFactory(logFactory LogFactory, metrics Metrics) RouterFactory {
	return &routerFactoryImpl{
		logFactory: logFactory,
		metrics:    metrics,
	}
}

/* RouterFactory implementation */

func (r *routerFactoryImpl) NewRouter() Router {
	return &httpRouterImpl{
		logFactory: r.logFactory,
		metrics:    r.metrics,
		router:     httprouter.New(),
	}
}

/* Router implementation */

func (r *httpRouterImpl) Handle(method, path string, metaFunc MetaFunc, handle Handle) {
	wrappedHandle := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		ww := NewWrappedResponseWriter(w)
		u := HandlerUtils{
			MetaFunc:   metaFunc,
			LogFactory: r.logFactory,
			Metrics:    r.metrics,
			ParamsFunc: NewRouteParamsFunc(p),
		}

		handle(ww, req, u)
	}

	r.router.Handle(method, path, wrappedHandle)
}

func (r *httpRouterImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

/* HandlerUtils methods */

// NewLoggerWithMeta instantiates and returns a new logger containing the provided meta.
func (u HandlerUtils) NewLoggerWithMeta(meta map[string]string, r *http.Request) Logger {
	combinedMeta := combineMetas(u.MetaFunc(r, u.ParamsFunc), meta)
	return u.LogFactory.NewLogger(combinedMeta)
}

// NewRouteParamsFunc returns a function to return RouteParams based on httprouter.Params
func NewRouteParamsFunc(p httprouter.Params) RouteParamsFunc {
	return func() RouteParams {
		rp := make(map[string]string)

		for i := range p {
			rp[p[i].Key] = p[i].Value
		}
		return rp
	}
}
