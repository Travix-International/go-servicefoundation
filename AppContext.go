package servicefoundation

import (
	"time"

	metrics "github.com/Travix-International/go-metrics"
)

// AppContext is the basis for application contexts
type AppContext interface {
	Metrics() *metrics.Metrics
	Version() AppVersion
	SetMetrics(v *metrics.Metrics)
	SetVersion(v AppVersion)
}

type AppVersion struct {
	MainVersion string
	MinVersion  string
	GitHash     string
	BuildDate   time.Time
}

type ContextBase struct {
	AppContext
	metrics *metrics.Metrics
	version AppVersion
}

func (ctx *ContextBase) Metrics() *metrics.Metrics {
	return ctx.metrics
}

func (ctx *ContextBase) Version() AppVersion {
	return ctx.version
}

func (ctx *ContextBase) SetMetrics(v *metrics.Metrics) {
	ctx.metrics = v
}

func (ctx *ContextBase) SetVersion(v AppVersion) {
	ctx.version = v
}
