package main

import (
	"myweb/app/gateway/internal/handler"
)

type Handlers struct {
	userHandler *handler.UserHandler
}

func NewGateway(userHandler *handler.UserHandler) *Handlers {
	return &Handlers{
		userHandler: userHandler,
	}
}
