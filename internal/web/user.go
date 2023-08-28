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

type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp2.Regexp
	passwordExp *regexp2.Regexp
}

func NewUserHandler(userServer *service.UserService) *UserHandler {
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
	}
}

// RegisterRouter 注册路由
func (u *UserHandler) RegisterRouter(server *gin.Engine) {
	userRouter := server.Group("/users")
	userRouter.POST("/signup", u.SignUp)
	userRouter.POST("/login", u.LoginJWT)
	userRouter.PUT("/edit", u.Edit)
	userRouter.GET("/profile", u.ProfileJWT)

}

// Profile 测试权限信息
func (u *UserHandler) Profile(ctx *gin.Context) {
	ctx.String(http.StatusOK, "你看到了。。。")
}

// ProfileJWT 测试权限信息
func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
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

	ctx.String(http.StatusOK, fmt.Sprintf("%s看到了。。。", claims.Email))
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

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Email:     req.Email,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	jwtToken, _ := token.SignedString([]byte("tbkykLFqpai8IwdLt9N20HfAs FZoK1uA"))
	ctx.Header("X-jwt-token", jwtToken)
	fmt.Printf("%v\n", result)
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

// Edit 更新信息
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
	Email     string
	UserAgent string
}
