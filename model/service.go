package model

import (
	"context"
	"time"
)

type (
	ShutdownFunc func(log Logger)

	ServiceGlobals struct {
		AppName           string
		ServerName        string
		DeployEnvironment string
		VersionNumber     string
	}

	ServiceOptions struct {
		Globals               ServiceGlobals
		Port                  int
		ReadinessPort         int
		InternalPort          int
		Logger                Logger
		Metrics               Metrics
		RouterFactory         RouterFactory
		ServiceHandlerFactory ServiceHandlerFactory
		VersionBuilder        VersionBuilder
		ShutdownFunc          ShutdownFunc
		ExitFunc              ExitFunc
		ServerTimeout         time.Duration
	}

	Service interface {
		Run(ctx context.Context)
		AddRoute(name string, routes []string, methods []string, middlewares []Middleware, handler Handle)
	}
)
