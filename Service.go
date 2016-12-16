package servicefoundation

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	ServiceSettings struct {
		Name         string
		CorsOrigins  []string
		HttpPort     int
		LogMinFilter string
	}

	ErrorResponse struct {
		Message string
	}

	RoutesDefinitionFunc func(mux *http.ServeMux)
)

func Run(settings ServiceSettings, routeDefinitionFunc RoutesDefinitionFunc, readiness ContextHandler, ctx AppContext) {
	svcSettings = settings
	routesFn = routeDefinitionFunc
	readinessHandler = readiness
	appCtx = ctx

	runReadinessServer()
	runInternalServer()
	runServer()
}

func runReadinessServer() {
	const subsystem = "readiness_service"

	port := fmt.Sprintf(":%v", svcSettings.HttpPort+readinessPortOffset)

	mux := http.NewServeMux()
	mux.HandleFunc("/", HistogramHandler(subsystem, "root", appCtx, RootHandler(subsystem, appCtx)))
	mux.HandleFunc("/service/liveness", CounterHandler(subsystem, "liveness", appCtx, LivenessHandler))
	mux.HandleFunc("/service/readiness", HistogramHandler(subsystem, "readiness", appCtx, readinessHandler(appCtx)))

	log.Printf("%v %v running on localhost%v.", svcSettings.Name, subsystem, port)
	go http.ListenAndServe(port, mux)
}

func runInternalServer() {
	const subsystem = "internal_service"

	port := fmt.Sprintf(":%v", svcSettings.HttpPort+internalPortOffset)

	mux := http.NewServeMux()
	mux.HandleFunc("/", HistogramHandler(subsystem, "root", appCtx, RootHandler(subsystem, appCtx)))
	mux.HandleFunc("/health_check", CounterHandler(subsystem, "health_check", appCtx, DontCacheHandler(HealthHandler)))
	mux.HandleFunc("/metrics", CounterHandler(subsystem, "metrics", appCtx, MetricsHandler))
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
