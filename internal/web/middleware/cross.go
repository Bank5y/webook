package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type CrossMiddlewareBuilder struct {
}

func NewCrossMiddlewareBuilder() *CrossMiddlewareBuilder {
	return &CrossMiddlewareBuilder{}
}

func (c *CrossMiddlewareBuilder) Build() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"https://localhost"},
		//传递信息跨域
		ExposeHeaders:    []string{"X-ijwt-token", "X-refresh-token"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
}
