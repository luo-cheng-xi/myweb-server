//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"
	"myweb/app/gateway/internal/handler"
	"myweb/app/gateway/pkg/log"
	"myweb/app/gateway/rpc"
)

func InitGateway() (*Handlers, error) {
	wire.Build(
		NewGateway,
		handler.NewUserHandler,
		rpc.NewUserConn,
		log.NewLogger,
	)
	return &Handlers{}, nil
}
