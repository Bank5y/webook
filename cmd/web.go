package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"webook/internal/web"
	"webook/internal/web/middleware"
)

func InitGin(middlewares []gin.HandlerFunc, hdl *web.UserHandler) *gin.Engine {
	engine := gin.Default()
	//初始化中间件
	engine.Use(middlewares...)
	hdl.RegisterRouter(engine)
	return engine
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.NewCrossMiddlewareBuilder().Build(),
		middleware.NewLoginJWTMiddlewareBuilder().
			Ignore("/users/login").
			Ignore("/users/login_sms/code/send").
			Ignore("/users/lo gin_sms").
			Ignore("/users/signup").
			Build(),
		sessions.Sessions("mySessions",
			memstore.NewStore(
				[]byte("tbkykLFqpai8IwdLt9N20HfAsFZoK1uA"),
				[]byte("Gv08GPb5tXIjrtQ8m2cwAVukIkUkDBLG"),
			),
		),
	}
}