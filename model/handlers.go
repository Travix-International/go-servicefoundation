package model

import "github.com/julienschmidt/httprouter"

type (
	ExitFunc func(int)

	ServiceHandlerFactory interface {
		WrapHandler(string, string, []Middleware, Handle) httprouter.Handle
		CreateRootHandler() Handle
		CreateReadinessHandler() Handle
		CreateLivenessHandler() Handle
		CreateQuitHandler() Handle
		CreateHealthHandler() Handle
		CreateVersionHandler() Handle
		CreateMetricsHandler() Handle
	}
)
