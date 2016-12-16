package servicefoundation

import (
	"time"

	metrics "github.com/Travix-International/go-metrics"
	"github.com/Travix-International/logger"
)

// AppContext is the basis for application contexts
type AppContext interface {
	Metrics() *metrics.Metrics
	Version() AppVersion
	Logger() *logger.Logger
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
	logger  *logger.Logger
	version AppVersion
}

func (ctx *ContextBase) Metrics() *metrics.Metrics {
	return ctx.metrics
}

func (ctx *ContextBase) Version() AppVersion {
	return ctx.version
}

func (ctx *ContextBase) Logger() *logger.Logger {
	return ctx.logger
}

func (ctx *ContextBase) SetMetrics(v *metrics.Metrics) {
	ctx.metrics = v
}

func (ctx *ContextBase) SetVersion(v AppVersion) {
	ctx.version = v
}

func (ctx *ContextBase) SetLogger(l *logger.Logger) {
	ctx.logger = l
}
