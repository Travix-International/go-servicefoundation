package servicefoundation

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/rs/cors"
)

const (
	ACCEPT_HEADER       = "Accept"
	ACCEPT_JSON         = "application/json"
	ACCEPT_XML          = "application/xml"
	readinessPortOffset = 1
	internalPortOffset  = 2
)

var (
	svcSettings      ServiceSettings
	appCtx           AppContext
	routesFn         RoutesDefinitionFunc
	readinessHandler ContextHandler
)

type (
	// ServiceSettings is the set of properties needed to run the base set of http services
	ServiceSettings struct {
		Name         string
		CorsOptions  *cors.Options
		HttpPort     int
		LogMinFilter string
	}

	ErrorResponse struct {
		Message string
	}

	// RoutesDefinitionFunc is used to communicate a set of custom HTTP routes to the service foundation
	RoutesDefinitionFunc func(mux *http.ServeMux)
)

// Run starts the service, which includes the readiness http service, internal http service and the application-specific
// http service, which typically includes a set of custom routes
func Run(settings ServiceSettings, routeDefinitionFunc RoutesDefinitionFunc, readiness ContextHandler, ctx AppContext) {
	svcSettings = settings
	routesFn = routeDefinitionFunc
	readinessHandler = readiness
	appCtx = ctx

	RunReadinessServer()
	RunInternalServer()
	runServer()
}

// ConfigureService can be used to configure the properties of the services. There's no need to call this from your application
// since Run will take care of that. But, this can be useful in cases where you do not want to use Run, for example to run
// only parts of the service foundation.
func ConfigureService(settings ServiceSettings, routeDefinitionFunc RoutesDefinitionFunc, readiness ContextHandler, ctx AppContext) {
}

// RunReadinessServer runs the readiness service as a go-routine
func RunReadinessServer() {
	const subsystem = "readiness_service"

	port := fmt.Sprintf(":%v", svcSettings.HttpPort+readinessPortOffset)

	mux := http.NewServeMux()
	mux.HandleFunc("/", HistogramHandler(subsystem, "root", appCtx, RootHandler(subsystem, appCtx)))
	mux.HandleFunc("/service/liveness", CounterHandler(subsystem, "liveness", appCtx, DontCacheHandler(LivenessHandler)))
	mux.HandleFunc("/service/readiness", HistogramHandler(subsystem, "readiness", appCtx, DontCacheHandler(readinessHandler(appCtx))))

	log.Printf("%v %v running on localhost%v.", svcSettings.Name, subsystem, port)
	go http.ListenAndServe(port, mux)
}

// RunInternalServer runs the readiness service as a go-routine
func RunInternalServer() {
	const subsystem = "internal_service"

	port := fmt.Sprintf(":%v", svcSettings.HttpPort+internalPortOffset)

	mux := http.NewServeMux()
	mux.HandleFunc("/", HistogramHandler(subsystem, "root", appCtx, RootHandler(subsystem, appCtx)))
	mux.HandleFunc("/health_check", CounterHandler(subsystem, "health_check", appCtx, DontCacheHandler(HealthHandler)))
	mux.HandleFunc("/metrics", CounterHandler(subsystem, "metrics", appCtx, DontCacheHandler(MetricsHandler)))
	mux.HandleFunc("/quit", DontCacheHandler(QuitHandler))

	log.Printf("%v %v running on localhost%v.", svcSettings.Name, subsystem, port)
	go http.ListenAndServe(port, mux)
}

func runServer() {
	const subsystem = "service"

	mux := http.NewServeMux()
	mux.HandleFunc("/service/version", CounterHandler(subsystem, "version", appCtx, VersionHandler(appCtx)))
	mux.HandleFunc("/service/liveness", CounterHandler(subsystem, "liveness", appCtx, LivenessHandler))
	//Note: This shouldn't be here, but for some reason Kubernetes requests readiness from the main port.
	mux.HandleFunc("/service/readiness", HistogramHandler(subsystem, "readiness", appCtx, readinessHandler(appCtx)))
	mux.HandleFunc("/", HistogramHandler(subsystem, "root", appCtx, RootHandler(subsystem, appCtx)))

	routesFn(mux)

	port := fmt.Sprintf(":%v", svcSettings.HttpPort)
	log.Printf("%v service running on localhost%v.", svcSettings.Name, port)

	http.ListenAndServe(port, mux)
}

func JSON(w http.ResponseWriter, statusCode int, content interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(content)
}

func XML(w http.ResponseWriter, statusCode int, content interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)

	w.Write([]byte(xml.Header))
	xml.NewEncoder(w).Encode(content)
}

func AcceptsXML(r *http.Request) bool {
	return !strings.Contains(r.Header.Get(ACCEPT_HEADER), ACCEPT_JSON) && strings.Contains(r.Header.Get(ACCEPT_HEADER), ACCEPT_XML)
}

func WriteResponse(w http.ResponseWriter, r *http.Request, statusCode int, content interface{}) {
	if AcceptsXML(r) {
		XML(w, statusCode, content)
		return
	}
	JSON(w, statusCode, content)
}

func SetCaching(w http.ResponseWriter, maxAge int) {
	w.Header().Set("Vary", "Accept, Origin") // Because we don't want to mix XML and JSON in the cache!
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%v", maxAge))
}
