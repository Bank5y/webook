package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
	"webook/internal/web"
)

// LoginJWTMiddlewareBuilder JWT 登录校验
type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
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
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokens := strings.Split(tokenHeader, " ")
		if len(tokens) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//jwt token处理
		tokenStr := tokens[1]
		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("tbkykLFqpai8IwdLt9N20HfAsFZoK1uA"), nil
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

		//jwt token 刷新
		now := time.Now()
		if claims.ExpiresAt.Sub(now) < time.Second*50 {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			tokenStr, err = token.SignedString([]byte("tbkykLFqpai8IwdLt9N20HfAsFZoK1uA"))
			if err != nil {
				//处理错误
			}
			ctx.Header("X-jwt-token", tokenStr)
		}
		ctx.Set("claims", claims)
	}
}
