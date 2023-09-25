package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"net/http"
	"webook/internal/web"
	"webook/internal/web/ijwt"
)

// LoginJWTMiddlewareBuilder JWT 登录校验
type LoginJWTMiddlewareBuilder struct {
	paths   []string
	cmd     redis.Cmdable
	handler ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(cmd redis.Cmdable, handler ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{cmd: cmd, handler: handler}
}

func (l *LoginJWTMiddlewareBuilder) Ignore(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		//jwt验证
		tokenStr := l.handler.ExtractToken(ctx)

		claims := &ijwt.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(web.JwtTokenKey), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//没登录
		if token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//验证Useragent
		if claims.UserAgent != ctx.Request.UserAgent() {
			//严重的安全问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		cnt, err := l.cmd.Exists(ctx, fmt.Sprintf("users:uuid:%s", claims.Ssid)).Result()
		if err != nil || cnt > 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return

		}

		ctx.Set("claims", claims)
	}
}
