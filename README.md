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

To do:
* De-normalize project structure (because Go)
* Standardize metrics
* Standardize log messages
* Add request/response logging as middleware
* Extend logging with meta information
* De-duplicate CORS elements in slices
* Automated documentation (GoDocs?)
* Code checks?
