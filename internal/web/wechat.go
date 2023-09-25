package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"time"
	"webook/internal/service"
	"webook/internal/service/oauth2/wechat"
)

type OAuthWechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	jwtHandler
	stateKey []byte
}

func NewOAuthWechatHandler(svc wechat.Service, userSvc service.UserService) *OAuthWechatHandler {
	return &OAuthWechatHandler{
		svc:      svc,
		userSvc:  userSvc,
		stateKey: []byte("tbkykLFqpai8IwdLt9N20HfAsFZoK1uA"),
	}
}

func (o *OAuthWechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oath2/wechat")
	g.GET("/authurl", o.AuthURL)
	g.Any("/callback", o.Callback)

}

func (o *OAuthWechatHandler) AuthURL(ctx *gin.Context) {
	state := uuid.New()
	url, err := o.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: "5",
			Msg:  "构造扫码登录失败",
		})
	}
	err = o.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: "5",
			Msg:  "系统异常",
			Data: nil,
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Data: url,
	})
	return
}

func (o *OAuthWechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 3)),
		},
	})
	tokenStr, err := token.SignedString(o.stateKey)
	if err != nil {
		return err
	}
	ctx.SetCookie("ijwt-state", tokenStr, 600, "/oath2/wechat/callback", "", false, true)
	return nil
}

func (o *OAuthWechatHandler) Callback(ctx *gin.Context) {
	//验证code
	code := ctx.Query("code")
	err := o.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: "5",
			Msg:  "登录失败",
			Data: nil,
		})
		return
	}
	info, err := o.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: "5",
			Msg:  "登录失败",
			Data: nil,
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: "5",
			Msg:  "系统错误",
			Data: nil,
		})
		return
	}
	u, err := o.userSvc.FindOrCreateByWechat(ctx, info)

	err = o.setLoginToken(ctx, u.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: "5",
			Msg:  "系统错误",
			Data: nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: "3",
		Msg:  "验证成功",
		Data: nil,
	})
}

func (o *OAuthWechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	cookie, err := ctx.Cookie("ijwt-state")
	if err != nil {
		return fmt.Errorf("拿不到state的cookie,%w", err)
	}
	var sc StateClaims
	token, err := jwt.ParseWithClaims(cookie, &sc, func(token *jwt.Token) (interface{}, error) {
		return o.stateKey, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("token已经过期,%w", err)

	}
	if sc.State != state {
		return fmt.Errorf("state不相等,%w", err)
	}
	return err
}

type StateClaims struct {
	State string
	jwt.RegisteredClaims
}
