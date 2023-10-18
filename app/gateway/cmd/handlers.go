package main

import (
	"myweb/app/gateway/internal/handler"
	"myweb/app/gateway/register"
)

type Handlers struct {
	userHandler *handler.UserHandler
	register    *register.EtcdRegister
}

func NewGateway(userHandler *handler.UserHandler, etcdRegister *register.EtcdRegister) *Handlers {
	return &Handlers{
		register:    etcdRegister,
		userHandler: userHandler,
	}
}
