//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package server

import (
	"github.com/google/wire"
	"myweb/app/user/internal/cache"
	"myweb/app/user/internal/conf"
	"myweb/app/user/internal/dao"
	"myweb/app/user/internal/logic"
	"myweb/app/user/pkg/log"
	"myweb/app/user/pkg/util"
)

func InitUserServer() (*UserServer, error) {
	wire.Build(
		NewUserServer,
		logic.NewUserLogic,
		dao.NewUserDao,
		dao.NewData,
		cache.NewCache,
		util.ProviderSet,
		log.NewLogger,
		conf.NewConfig,
	)
	return &UserServer{}, nil
}
