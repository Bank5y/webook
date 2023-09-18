package web

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/dao"
	"webook/internal/service"
	svcmocks "webook/internal/service/mocks"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) service.UserService
		reqBody string

		wantCode int
		wantBody string
	}{
		//正常流畅
		{
			name: "SignUp-success",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "25946185542@qq.com",
					Password: "1234qwe56asd@",
				})
				return userService
			},
			reqBody: `
{
    "email": "25946185542@qq.com",
    "password": "1234qwe56asd@"
}
`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		//错误参数
		{
			name: "SignUp-varError",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userService := svcmocks.NewMockUserService(ctrl)
				return userService
			},
			reqBody: `
{
    "email": "25946185542@qq.com",
}
`,
			wantCode: http.StatusBadRequest,
		},
		//邮箱有误
		{
			name: "SignUp-emailError",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userService := svcmocks.NewMockUserService(ctrl)
				return userService
			},
			reqBody: `
{
    "email": "25946185542qq.com",
    "password": "1234qwe56asd@"
}
`,
			wantCode: http.StatusOK,
			wantBody: "邮箱有误",
		},
		//密码有误
		{
			name: "Signup-passwordError",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userService := svcmocks.NewMockUserService(ctrl)
				return userService
			},
			reqBody: `
{
    "email": "25946185542@qq.com",
    "password": "1234q"
}
`,
			wantCode: http.StatusOK,
			wantBody: "密码格式错误！",
		},
		//邮箱冲突
		{
			name: "Sign-emailDupError",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "2594618554@qq.com",
					Password: "1234qwe56asd@",
				}).Return(dao.ErrUserDuplicateEmail)
				return userService
			},
			reqBody: `
{
    "email": "2594618554@qq.com",
    "password": "1234qwe56asd@"
}
`,
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			userHandler := NewUserHandler(tc.mock(ctrl), nil)
			userHandler.RegisterRouter(server)

			req := httptest.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())

		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService)

		reqBody string

		wantResult Result
	}{
		//成功流程
		{
			name: "LoginSMS-success",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "8188181", "866108").
					Return(true, nil)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), domain.User{Phone: "8188181"}).
					Return(domain.User{
						Id:       1,
						Email:    "",
						Phone:    "8188181",
						Password: "",
					}, nil)
				return userSvc, codeSvc
			},
			reqBody: `{
"phone":"8188181",
"code":"866108"
}`,
			wantResult: Result{
				Code: "510002",
				Msg:  "验证成功",
			},
		},
		//绑定不成功 系统错误
		{
			name: "LoginSMS-sysError",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, codeSvc
			},
			reqBody: `{
phone:"8188181,
}`,
			wantResult: Result{
				Code: "510002",
				Msg:  "系统错误",
				Data: "invalid character 'p' looking for beginning of object key string",
			},
		},
		//验证不成功 系统错误
		{
			name: "LoginSMS-verError",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "8188181", "866108").
					Return(true, service.ErrCodeSendTooMany)
				return userSvc, codeSvc
			},
			reqBody: `{
"phone":"8188181",
"code":"866108"
}`,
			wantResult: Result{
				Code: "510002",
				Msg:  "系统错误",
				Data: "发送太频繁",
			},
		},
		//验证不成功 验证失败
		{
			name: "LoginSMS-success",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "8188181", "866108").
					Return(false, nil)
				return userSvc, codeSvc
			},
			reqBody: `{
"phone":"8188181",
"code":"866108"
}`,
			wantResult: Result{
				Code: "510002",
				Msg:  "验证失败",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctrl.Finish()

			userHandler := NewUserHandler(tc.mock(ctrl))
			server := gin.Default()
			userHandler.RegisterRouter(server)
			req := httptest.NewRequest(http.MethodPost, "/users/login_sms",
				bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "")

			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)
			val, err := json.Marshal(tc.wantResult)
			require.NoError(t, err)
			assert.Equal(t, string(val), resp.Body.String())
		})
	}
}

func TestSetJwtToken(t *testing.T) {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			//不好测
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		UserId:    1,
		UserAgent: "",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	jwtToken, err := token.SignedString([]byte(JwtTokenKey))
	require.NoError(t, err)
	println(jwtToken)
}
