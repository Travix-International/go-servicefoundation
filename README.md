# ServiceFoundation [![Build Status](https://travis-ci.org/Travix-International/go-servicefoundation.svg?branch=master)](https://travis-ci.org/Travix-International/go-servicefoundation?branch=master)

[![Go Report Card](https://goreportcard.com/badge/github.com/Travix-International/go-servicefoundation)](https://goreportcard.com/report/github.com/Travix-International/go-servicefoundation) [![Coverage Status](https://coveralls.io/repos/github/Travix-International/go-servicefoundation/badge.svg?branch=master)](https://coveralls.io/github/Travix-International/go-servicefoundation?branch=master) 
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/Travix-International/go-servicefoundation/blob/master/LICENSE)

> Create new Web Services using convention-based configuration.

More documentation to be found on [GoDoc](https://godoc.org/github.com/Travix-International/go-servicefoundation).

ServiceFoundation enables you to create Web Services containing:

* 3 access levels (public, readiness and internal)
* Customizable logging (defaults to go-logger)
* Customizable metrics collection (defaults to go-metrics)
* Out-of-the-box middleware for panic handling, no-cache, counters, histograms and CORS.
* Default and overridable handling of catch-all (root), liveness, health, version and readiness 
* Handling of SIGTERM and SIGINT with a custom shutdown function to properly free your own resources.
* Customizable server timeouts
* Request/response logging as middleware
* Support service warm-up through state customization
* Standardized metrics (defaults to go-metrics)
* Standardized log messages in JSON format
* Adding route-specific meta fields to log messages
* Default handling of pre-flight requests
* Optional authorization middleware

To do:
- [ ] Automated documentation (GoDocs?)

## Package usage

Include this package into your project with:

```
go get github.com/Travix-International/go-servicefoundation
```

Although all components can be extended, the easiest way to use ServiceFoundation is to use the boilerplate version:

```go
package main

import (
	"context"
	"net/http"

	sf "github.com/Travix-International/go-servicefoundation"
)

// Set these variables at runtime using -ldflags
var gitHash, versionNumber, buildDate string

func main() {
	opt := sf.NewDefaultServiceOptions("AppGroup","HelloWorldService")
	svc := sf.NewService(opt)

	svc.AddRoute(
		"helloworld",
		[]string{"/helloworld"},
		sf.MethodsForGet,
		[]sf.Middleware{sf.PanicTo500, sf.CORS, sf.RequestMetrics},
		func(r *http.Request, _ sf.RouterParams) map[string]string {
        		return make(map[string]string)
        },
		func(w sf.WrappedResponseWriter, r *http.Request, _ sf.HandlerUtils) {
			w.WriteResponse(r, http.StatusOK, "hello world!")
		})

	svc.Run(context.Background()) // blocks execution
}
```

The following environment variables are used by ServiceFoundation:

|Name              |Used for                                                  
|------------------|----------------------------------------------------------
|CORS_ORIGINS      |Comma-separated list of CORS origins (default:*)          
|HTTPPORT          |Port used for exposing the public endpoint (default: 8080)
|LOG_MINFILTER     |Minimum filter for log writing (default: Warning)         
|APP_NAME          |Name of the application (HelloWorldService)               
|SERVER_NAME       |Name of the server instance (helloworldservice-1234)      
|DEPLOY_ENVIRONMENT|Name of the deployment environment (default: staging)     

## Dependencies

Although ServiceFoundation contains interfaces to hide any external dependencies, the default configuration depends 
on the following packages:

* [github.com/travix-International/logger](https://github.com/travix-International/logger)
* [github.com/Travix-International/go-metrics](https://github.com/Travix-International/go-metrics)
* [github.com/julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)
* [github.com/rs/cors](https://github.com/rs/cors)
* [github.com/prometheus/client_golang/prometheus/promhttp](https://github.com/prometheus/prometheus)


## Extending ServiceFoundation

You can use the `servicefoundation.CreateService(ServiceOptions)` method for extension. The provided `ServiceOptions` 
struct contains all the things you can extend. If you want to create your own `ReadinessHandler`, you can do so as 
follows:

```go
package main

import (
	"context"
	"net/http"
	"time"

	sf "github.com/Travix-International/go-servicefoundation"
)

var gitHash, versionNumber, buildDate string

type CustomServiceStateManager struct {
	sf.ServiceStateManager
	isWarmedUp bool
}

func (r *CustomServiceStateManager) IsLive() bool {
	return true
}

func (r *CustomServiceStateManager) IsReady() bool {
	return r.isWarmedUp
}

func (r *CustomServiceStateManager) IsHealthy() bool {
	return true
}

func (r *CustomServiceStateManager) WarmUp() {
	// This routine is not run in a go-routine, so we could use it to some last-minute
	// bindings. If we do want to run it asynchronously, we have to use a go-routine here.
	go func() {
        // Simulating warm-up time...
        time.Sleep(10 * time.Second)
        r.isWarmedUp = true
    }()
}

func (r *CustomServiceStateManager) ShutDown(logger sf.Logger) {
    logger.Info("GracefulShutdown", "Handling graceful shutdown")
}

func main() {
    // Use a global meta for logging additional fields during the service lifecycle
    globalMeta := make(map[string]string)
    globalMeta["hello"] = "world"
    
	opt := sf.NewServiceOptions(
		"AppGroup", "HelloWorldService",
		sf.MethodsForGet,
		sf.BuildVersion{
			GitHash:       gitHash,
			VersionNumber: versionNumber,
			BuildDate:     buildDate,
		}, globalMeta)
	
	opt.AuthFunc = func(_ sf.WrappedResponseWriter, _ *http.Request, u sf.HandlerUtils) bool {
        // Implement your own authorization here
        log := u.LogFactory.NewLogger(make(map[string]string))
        log.Info("Authorize", "Authorization requested")
        return true 
    }
	
	// Use this in case you want to handle the public root endpoint yourself instead of relying 
	// on the default catch-all handling.
	opt.UsePublicRootHandler = true

	opt.SetState(&CustomServiceStateManager{})

	svc := sf.NewService(opt) // Also triggers the service WarmUp

	svc.AddRoute(
		"helloworld",
		[]string{"/helloworld"},
		sf.MethodsForGet,
		[]sf.Middleware{sf.PanicTo500, sf.CORS, sf.RequestMetrics},
		func(r *http.Request, _ sf.RouterParams) map[string]string {
            // Use a route-specific meta to log additional fields for handling a route request 
            routeMeta := make(map[string]string)
            routeMeta["hello"] = "route"
			return routeMeta 
        },
		func(w sf.WrappedResponseWriter, r *http.Request, u sf.HandlerUtils) {
            log := u.LogFactory.NewLogger(make(map[string]string))
			log.Info("HelloWorld", "Endpoint called")
			w.WriteResponse(r, http.StatusOK, "hello world!")
		})

	svc.Run(context.Background()) // blocks execution
}
```

[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/Travix-International/go-servicefoundation/blob/master/LICENSE)
