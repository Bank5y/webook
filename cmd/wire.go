//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//基础组件
		ioc.InitDB, ioc.InitRedis,
		//cache dao
		cache.NewCodeLocalCache, cache.NewUserRedisCache, dao.NewUserDao,
		//repo
		wire.NewSet(repository.NewCodeCacheRepository, repository.NewUserCacheRepository,
			wire.Bind(new(repository.CodeRepository), new(*repository.CodeCacheRepository)),
			wire.Bind(new(repository.UserRepository), new(*repository.UserCacheRepository)),
		),
		//service
		wire.NewSet(service.NewCodeDevService, service.NewUserDevService,
			wire.Bind(new(service.CodeService), new(*service.CodeDevService)),
			wire.Bind(new(service.UserService), new(*service.UserDevService)),
		),
		ioc.InitSMSService,
		//web
		web.NewUserHandler, ioc.InitGin, ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
