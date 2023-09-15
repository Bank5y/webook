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
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//基础组件
		InitDB, InitRedis,
		//cache dao
		cache.NewCodeLocalCache, cache.NewUserRedisCache, dao.NewUserDao,
		//repo
		repository.NewCodeCacheRepository, repository.NewUserCacheRepository,
		//service
		service.NewCodeDevService, service.NewUserDevService, InitSMSService,
		//web
		web.NewUserHandler, InitGin, InitMiddlewares,
	)
	return new(gin.Engine)
}
