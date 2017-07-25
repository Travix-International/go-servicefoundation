package site

import (
	"github.com/Prutswonder/go-servicefoundation/model"
	"github.com/julienschmidt/httprouter"
)

type routerFactoryImpl struct {
}

func (r *routerFactoryImpl) CreateRouter() *httprouter.Router {
	return httprouter.New()
}

func CreateRouterFactory() model.RouterFactory {
	return &routerFactoryImpl{}
}
