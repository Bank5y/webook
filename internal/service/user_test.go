package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
)

func TestUserDevService_Login(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		//输入
		context   context.Context
		inputUser domain.User

		//预期
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "正常测试",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepository := repomocks.NewMockUserRepository(ctrl)
				userRepository.EXPECT().FindByEmail(gomock.Any(), domain.User{
					Email:    "12345@qq.com",
					Password: "5123412312asd@",
				}).
					Return(domain.User{
						Email:    "12345@qq.com",
						Phone:    "186xxx",
						Password: "$2a$10$e.u5gkPeXdL6s8tNpBcjSe1DPfHgZEL1jJ4kNoMuzxkVOzbeRRb9u2",
					}, nil)
				return userRepository
			},
			inputUser: domain.User{
				Email:    "12345@qq.com",
				Password: "5123412312asd@",
			},
			wantUser: domain.User{
				Email:    "12345@qq.com",
				Phone:    "186xxx",
				Password: "$2a$10$e.u5gkPeXdL6s8tNpBcjSe1DPfHgZEL1jJ4kNoMuzxkVOzbeRRb9u2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserDevService(tc.mock(ctrl))
			u, err := svc.Login(context.Background(), tc.inputUser)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)

		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("5123412312asd@"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}
