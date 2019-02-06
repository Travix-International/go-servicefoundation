package servicefoundation

type (
	// ServiceStateReader contains state methods used by the service's handler implementations.
	ServiceStateReader interface {
		IsLive() bool
		IsReady() bool
		IsHealthy() bool
	}

	// ServiceStateManager gives the service control over the warm-up and shutdown process.
	ServiceStateManager interface {
		ServiceStateReader
		WarmUp()
		ShutDown(Logger)
	}

	defaultServiceStateManagerImpl struct {
	}
)

// NewDefaultServiceStateManger instantiates a new basic ServiceStateManagerr implementation, which always returns true
// for it's state methods.
func NewDefaultServiceStateManger() ServiceStateManager {
	return &defaultServiceStateManagerImpl{}
}

/* Default ServiceStateManager implementation */

func (s *defaultServiceStateManagerImpl) IsLive() bool {
	// The default state manager is always live
	return true
}

func (s *defaultServiceStateManagerImpl) IsReady() bool {
	// The default state manager is ready immediately
	return true
}

func (s *defaultServiceStateManagerImpl) IsHealthy() bool {
	// The default state manager is always healthy
	return true
}

func (s *defaultServiceStateManagerImpl) WarmUp() {
	// The default state manager doesn't need to warm up anything
}

func (s *defaultServiceStateManagerImpl) ShutDown(logger Logger) {
	// The default state manager doesn't need to shut down anything
}
