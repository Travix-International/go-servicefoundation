package servicefoundation_test

import (
	metrics "github.com/Travix-International/go-metrics"
	servicefoundation "github.com/Travix-International/go-servicefoundation"
	"github.com/Travix-International/logger"
)

type UnitTestAppContext struct {
	servicefoundation.ContextBase
}

func NewAppContext(metricsNamespace string, version servicefoundation.AppVersion, loggy *logger.Logger) servicefoundation.AppContext {
	ctx := &UnitTestAppContext{}
	ctx.SetLogger(loggy)
	ctx.SetMetrics(metrics.NewMetrics(metricsNamespace, loggy))
	ctx.SetVersion(version)
	appContext := servicefoundation.AppContext(ctx)
	return appContext
}
