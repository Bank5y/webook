package web

import (
	"errors"
	"fmt"
	"github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web/ijwt"
)

const (
	biz = "login"
)

type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp2.Regexp
	passwordExp *regexp2.Regexp
	cmd         redis.Cmdable
	handler     ijwt.Handler
}

func NewUserHandler(userServer service.UserService, codeSvc service.CodeService, cmd redis.Cmdable, jwtHandler ijwt.Handler) *UserHandler {
	const (
		//email regex
		emailRegexPattern = `^[A-Za-z0-9\u4e00-\u9fa5]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
		//password regex
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	emailExp := regexp2.MustCompile(emailRegexPattern, regexp2.None)
	passwordExp := regexp2.MustCompile(passwordRegexPattern, regexp2.None)
	return &UserHandler{
		svc:         userServer,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		codeSvc:     codeSvc,
		cmd:         cmd,
		handler:     jwtHandler,
	}
}

// RegisterRouter 注册路由
func (u *UserHandler) RegisterRouter(server *gin.Engine) {
	userRouter := server.Group("/users")
	userRouter.POST("/signup", u.SignUp)
	userRouter.POST("/login", u.LoginJWT)
	userRouter.PUT("/edit", u.Edit)
	userRouter.GET("/profile", u.ProfileJWT)
	userRouter.POST("/login_sms/code/send", u.SendLoginSMSCode)
	userRouter.POST("/login_sms", u.LoginSMS)
	userRouter.POST("/logout", u.Logout)
	userRouter.POST("/refresh_token", u.RefreshToken)
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	ctx.Header("X-jwt-token", "")
	ctx.Header("X-refresh-token", "")
	claims, ok := ctx.MustGet("claims").(*ijwt.UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}
	err := u.cmd.Set(ctx, fmt.Sprintf("users:uuid:%s", claims.Ssid), "", time.Hour*7*24).Err()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "退出登录失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "退出登录",
	})
	return

}

// SendLoginSMSCode 发送验证码
func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	err = u.codeSvc.Send(ctx, biz, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送太频繁,请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: "5",
			Msg:  "系统错误",
		})
	}
	return
}

// Profile 测试权限信息
func (u *UserHandler) Profile(ctx *gin.Context) {
	ctx.String(http.StatusOK, "你看到了。。。")
}

// ProfileJWT 测试权限信息
func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	//可以断言必然有claims
	claims := ctx.MustGet("claims").(*ijwt.UserClaims)

	ctx.String(http.StatusOK, fmt.Sprintf("%d看到了。。。", claims.UserId))
}

// SignUp 注册
func (u *UserHandler) SignUp(ctx *gin.Context) {
	//接受json请求体
	type SignUpReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	//绑定请求体到结构体
	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	//邮箱正则验证
	ok, err := u.emailExp.MatchString(req.Email)
	println(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱有误")
		return
	}

	//密码正则验证
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码格式错误！")
		return
	}

	//DAO层数据处理
	err = u.svc.SignUp(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, dao.ErrUserDuplicateEmail) {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}

// Login 登录
func (u *UserHandler) Login(ctx *gin.Context) {
	type UserReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req UserReq
	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	result, err := u.svc.Login(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if errors.Is(err, service.ErrUserNotFind) {
		ctx.String(http.StatusOK, "用户名或密码错误！")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	//设置session
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: 10,
	})
	sess.Set("LoginSess", result.Email)
	err = sess.Save()
	if err != nil {
		return
	}

	ctx.String(http.StatusOK, "登录成功！")
	return
}

// LoginSMS 验证码登录
func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: "510002",
			Msg:  "系统错误",
			Data: err.Error(),
		})
		return
	}
	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: "510002",
			Msg:  "系统错误",
			Data: err.Error(),
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: "510002",
			Msg:  "验证失败",
		})
		return
	}
	user, err := u.svc.FindOrCreate(ctx, domain.User{
		Phone: req.Phone,
	})
	if err != nil {
		return
	}
	err = u.handler.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: "5",
			Msg:  "系统错误",
			Data: nil,
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: "510002",
		Msg:  "验证成功",
	})
	return
}

// LoginJWT JWT登录
func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type UserReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req UserReq
	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	result, err := u.svc.Login(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if errors.Is(err, service.ErrUserNotFind) {
		ctx.String(http.StatusOK, "用户名或密码错误！")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	err = u.handler.SetLoginToken(ctx, result.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "登录成功！")
	return
}

// LogOut 登出
func (u *UserHandler) LogOut(ctx *gin.Context) {
	//设置session
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	err := sess.Save()
	if err != nil {
		return
	}
	ctx.String(http.StatusOK, "退出登录成功")
}

// Edit 更新信息(Session版)
func (u *UserHandler) Edit(ctx *gin.Context) {
	type UpdateReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req UpdateReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	sess := sessions.Default(ctx)
	email := sess.Get("LoginSess")
	if email.(string) != req.Email {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	//密码正则验证
	ok, err := u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码格式错误！")
		return
	}

	err = u.svc.EditUserPassword(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "修改成功！")
	return
}

// RefreshToken 拿长token
func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	//只有这个接口 拿出来的才是长token 其他地方都是短token
	refreshToken := u.handler.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &rc, func(token *jwt.Token) (interface{}, error) {
		return []byte(RefreshTokenKey), nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = u.handler.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	cnt, err := u.cmd.Exists(ctx, fmt.Sprintf("users:uuid:%s", rc.Ssid)).Result()
	if err != nil || cnt > 0 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: "3",
		Msg:  "刷新成功",
	})
}
