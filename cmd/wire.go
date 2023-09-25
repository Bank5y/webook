//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/service/oauth2/wechat"
	"webook/internal/web"
	"webook/internal/web/ijwt"
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
			//绑定repo接口
			wire.Bind(new(repository.CodeRepository), new(*repository.CodeCacheRepository)),
			wire.Bind(new(repository.UserRepository), new(*repository.UserCacheRepository)),
		),
		//service
		wire.NewSet(service.NewCodeDevService, service.NewUserDevService, ioc.InitWechatService,
			//绑定service接口
			wire.Bind(new(wechat.Service), new(*wechat.DevService)),
			wire.Bind(new(service.CodeService), new(*service.CodeDevService)),
			wire.Bind(new(service.UserService), new(*service.UserDevService)),
		),
		ioc.InitSMSService,
		//web
		ioc.InitGin, ioc.InitMiddlewares, web.NewOAuthWechatHandler,
		wire.NewSet(web.NewUserHandler, ijwt.NewRedisJwt,
			wire.Bind(new(ijwt.Handler), new(*ijwt.RedisJwt)),
		),
	)
	return new(gin.Engine)
}
