package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) Ignore(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		sess := sessions.Default(ctx)
		email := sess.Get("LoginSess")
		if email == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now().UnixMilli()
		updateTime := sess.Get("update_time")
		updateTimeVal, _ := updateTime.(int64)
		if updateTime == nil || (now-updateTimeVal) > 2*1000 {
			sess.Set("update_time", now)
			sess.Options(sessions.Options{
				MaxAge: 10,
			})
		}
		err := sess.Save()
		if err != nil {
			panic(err)
		}
	}
}
