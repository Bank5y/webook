package ijwt

import "github.com/gin-gonic/gin"

type Handler interface {
	ExtractToken(ctx *gin.Context) string
	SetLoginToken(ctx *gin.Context, userId int) error
	ClearToken(ctx *gin.Context) error
	CheckSession(ctx *gin.Context, ssid string) error
	SetJWTToken(ctx *gin.Context, userId int, ssid string) error
	SetRefreshToken(ctx *gin.Context, userId int, ssid string) error
}
