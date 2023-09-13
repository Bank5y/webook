package web

import (
	"errors"
	"fmt"
	"github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/dao"
	"webook/internal/service"
)

const (
	biz = "login"
)

type UserHandler struct {
	svc         *service.UserService
	codeSvc     *service.CodeService
	emailExp    *regexp2.Regexp
	passwordExp *regexp2.Regexp
}

func NewUserHandler(userServer *service.UserService, codeSvc *service.CodeService) *UserHandler {
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
	println("111")
	//可以断言必然有claims
	c, ok := ctx.Get("claims")
	if !ok {
		//可以考虑监控住这里
		ctx.String(http.StatusOK, "系统错误")

	}
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}

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
			Data: nil,
		})
		return
	}
	user, err := u.svc.FindOrCreate(ctx, domain.User{
		Phone: req.Phone,
	})
	if err != nil {
		return
	}
	println(user.Id)
	err = u.setJWTToken(ctx, user.Id)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: "510002",
		Msg:  "验证成功",
		Data: nil,
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

	err = u.setJWTToken(ctx, result.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "登录成功！")
	return
}

// setJWTToken 设置JWTToken
func (u *UserHandler) setJWTToken(ctx *gin.Context, userId int) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		UserId:    userId,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	jwtToken, err := token.SignedString([]byte(JwtTokenKey))
	if err != nil {
		return err
	}
	ctx.Header("X-jwt-token", jwtToken)
	return nil
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

type UserClaims struct {
	jwt.RegisteredClaims
	//自定义存入token的字段
	UserId    int
	UserAgent string
}
