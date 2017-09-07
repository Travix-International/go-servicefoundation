# ServiceFoundation
Create new Web Services using convention-based configurations.

ServiceFoundation offers you Web Services containing:

* 3 access levels (public, readiness and internal)
* Customizable logging (defaults to go-logger)
* Customizable metrics collection (defaults to go-metrics)
* Out-of-the-box middleware for panic handling, no-cache, counters, histograms and CORS.
* Default and overridable handling of catch-all (root), liveness, health, version and readiness 
* Handling of SIGTERM and SIGINT with a custom shutdown function to properly free your own resources.
* Customizable server timeouts
* Request/response logging as middleware

To do:
* De-normalize project structure (because Go)
* Standardize metrics
* Standardize log messages
* Extend logging with meta information
* De-duplicate CORS elements in slices
* Automated documentation (GoDocs?)
* Code checks?

[![Build Status](https://travis-ci.org/Prutswonder/go-servicefoundation.svg?branch=v2)](https://travis-ci.org/Prutswonder/go-servicefoundation)

[![Coverage Status](https://coveralls.io/repos/github/Prutswonder/go-servicefoundation/badge.svg?branch=v2)](https://coveralls.io/github/Prutswonder/go-servicefoundation?branch=v2)

[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/Prutswonder/go-servicefoundation/blob/master/LICENSE)
