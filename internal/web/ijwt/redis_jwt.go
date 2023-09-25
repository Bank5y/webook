package ijwt

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/internal/web"
)

type RedisJwt struct {
	cmd redis.Cmdable
}

func NewRedisJwt(cmd redis.Cmdable) *RedisJwt {
	return &RedisJwt{cmd: cmd}
}

func (r *RedisJwt) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	if tokenHeader == "" {
		return ""
	}
	tokens := strings.Split(tokenHeader, " ")
	if len(tokens) != 2 {
		return ""
	}
	return tokens[1]
}

func (r *RedisJwt) SetLoginToken(ctx *gin.Context, userId int) error {
	ssid := uuid.New().String()
	err := r.SetJWTToken(ctx, userId, ssid)
	if err != nil {
		return err
	}
	err = r.SetRefreshToken(ctx, userId, ssid)
	if err != nil {
		return err
	}
	return err
}

func (r *RedisJwt) ClearToken(ctx *gin.Context) error {
	ctx.Header("X-jwt-token", "")
	ctx.Header("X-refresh-token", "")

	claims := ctx.MustGet("claims").(*UserClaims)
	return r.cmd.Set(ctx, fmt.Sprintf("user:ssid:%s", claims.Ssid), "", time.Hour*7*24).Err()
}

func (r *RedisJwt) CheckSession(ctx *gin.Context, ssid string) error {
	//TODO implement me
	panic("implement me")
}

func (r *RedisJwt) SetJWTToken(ctx *gin.Context, userId int, ssid string) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Ssid:      ssid,
		UserId:    userId,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	jwtToken, err := token.SignedString([]byte(web.JwtTokenKey))
	if err != nil {
		return err
	}
	ctx.Header("X-jwt-token", jwtToken)
	return nil
}

func (r *RedisJwt) SetRefreshToken(ctx *gin.Context, userId int, ssid string) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Ssid: ssid,
		Uid:  userId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	jwtToken, err := token.SignedString([]byte(web.RefreshTokenKey))
	if err != nil {
		return err
	}
	ctx.Header("X-refresh-token", jwtToken)
	return nil
}

type UserClaims struct {
	jwt.RegisteredClaims
	//自定义存入token的字段
	UserId    int
	Ssid      string
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Ssid string
	Uid  int
}
