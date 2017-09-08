# ServiceFoundation [![Build Status](https://travis-ci.org/Prutswonder/go-servicefoundation.svg?branch=v2)](https://travis-ci.org/Prutswonder/go-servicefoundation?branch=v2)

[![Go Report Card](https://goreportcard.com/badge/github.com/Prutswonder/go-servicefoundation)](https://goreportcard.com/report/github.com/Prutswonder/go-servicefoundation) [![Coverage Status](https://coveralls.io/repos/github/Prutswonder/go-servicefoundation/badge.svg?branch=v2)](https://coveralls.io/github/Prutswonder/go-servicefoundation?branch=v2) 
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/Prutswonder/go-servicefoundation/blob/master/LICENSE)

> Create new Web Services using convention-based configuration.

More documentation to be found on [GodDoc](https://godoc.org/github.com/Prutswonder/go-servicefoundation).

ServiceFoundation enables you to create Web Services containing:

* 3 access levels (public, readiness and internal)
* Customizable logging (defaults to go-logger)
* Customizable metrics collection (defaults to go-metrics)
* Out-of-the-box middleware for panic handling, no-cache, counters, histograms and CORS.
* Default and overridable handling of catch-all (root), liveness, health, version and readiness 
* Handling of SIGTERM and SIGINT with a custom shutdown function to properly free your own resources.
* Customizable server timeouts
* Request/response logging as middleware

To do:
- [ ] De-normalize project structure (because Go)
- [ ] Test extensibility 
- [ ] Standardize metrics
- [ ] Standardize log messages
- [ ] Extend logging with meta information
- [ ] Support service warm-up
- [ ] Overriding readiness 
- [ ] De-duplicate CORS elements in slices
- [ ] Automated documentation (GoDocs?)

## Package usage

Include this package into your project with:

```
go get github.com/Prutswonder/go-servicefoundation
```

Although all components can be extended, the easiest way to use ServiceFoundation is to use the boilerplate version:

```go
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/Prutswonder/go-servicefoundation"
	. "github.com/Prutswonder/go-servicefoundation/model"
)

func main() {
	svc := servicefoundation.CreateDefaultService("HelloWorldService", []string{http.MethodGet},
		func(log Logger) {
			log.Info("GracefulShutdown", "Handling graceful shutdown")
		})

	svc.AddRoute("helloworld", []string{"/helloworld"}, MethodsForGet, servicefoundation.DefaultMiddlewares,
		func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
			w.JSON(http.StatusOK, "hello world!")
		})

	svc.Run(context.Background()) // blocks execution
}
```

The following environment variables are used by ServiceFoundation:

|Name              |Used for                                                  |
|------------------|----------------------------------------------------------|
|CORS_ORIGINS      |Comma-separated list of CORS origins (default:*)          |
|HTTPPORT          |Port used for exposing the public endpoint (default: 8080)|
|LOG_MINFILTER     |Minimum filter for log writing (default: Warning)         |
|APP_NAME          |Name of the application (HelloWorldService)               |
|SERVER_NAME       |Name of the server instance (helloworldservice-1234)      |
|DEPLOY_ENVIRONMENT|Name of the deployment environment (default: staging)     |

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
struct contains all the things you can extend.


[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/Prutswonder/go-servicefoundation/blob/master/LICENSE)
