package syntoasy

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/service/sms"
	smsmocks "webook/internal/service/sms/syntoasy/mocks"
	pkgmocks "webook/pkg/mocks"
	"webook/pkg/ratelimit"
)

func TestTooManyErrFailoverService_Send(t *testing.T) {
	testCases := []struct {
		name string

		mock  func(ctrl *gomock.Controller) (limiter ratelimit.Limiter, service sms.Service)
		after func(service *TooManyErrFailoverService, err error) error

		wantIdx int64
		wantErr error
	}{
		//正常流程
		{
			name: "success-Send",
			mock: func(ctrl *gomock.Controller) (limiter ratelimit.Limiter, service sms.Service) {
				mockLimiter := pkgmocks.NewMockLimiter(ctrl)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				mockService := smsmocks.NewMockService(ctrl)
				mockService.EXPECT().Send(context.Background(), gomock.Any(), gomock.Any()).Return(nil)
				return mockLimiter, mockService
			},
			after: func(service *TooManyErrFailoverService, err error) error {
				return err
			},
			wantIdx: 0,
			wantErr: nil,
		},
		//限流出现问题
		{
			name: "success-limitErr",
			mock: func(ctrl *gomock.Controller) (limiter ratelimit.Limiter, service sms.Service) {
				mockLimiter := pkgmocks.NewMockLimiter(ctrl)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, errors.New("插件出现问题"))
				mockService := smsmocks.NewMockService(ctrl)
				return mockLimiter, mockService
			},
			after: func(service *TooManyErrFailoverService, err error) error {
				return err
			},
			wantIdx: 0,
			wantErr: errors.New("插件出现问题"),
		},
		//限流触发
		{
			name: "success-limitAct",
			mock: func(ctrl *gomock.Controller) (limiter ratelimit.Limiter, service sms.Service) {
				mockLimiter := pkgmocks.NewMockLimiter(ctrl)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				mockService := smsmocks.NewMockService(ctrl)
				return mockLimiter, mockService
			},
			after: func(service *TooManyErrFailoverService, err error) error {
				return err
			},
			wantIdx: 1,
			wantErr: nil,
		},
		//Send失败换svc
		{
			name: "success-Send",
			mock: func(ctrl *gomock.Controller) (limiter ratelimit.Limiter, service sms.Service) {
				mockLimiter := pkgmocks.NewMockLimiter(ctrl)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				mockService := smsmocks.NewMockService(ctrl)
				mockService.EXPECT().Send(context.Background(), gomock.Any(), gomock.Any()).Return(errors.New("send失败"))
				return mockLimiter, mockService
			},
			after: func(service *TooManyErrFailoverService, err error) error {
				return err
			},
			wantIdx: 1,
			wantErr: errors.New("send失败"),
		},
		//已经有respTime的正常流程
		{
			name: "success-RespTimeTooMore",

			mock: func(ctrl *gomock.Controller) (limiter ratelimit.Limiter, service sms.Service) {
				mockLimiter := pkgmocks.NewMockLimiter(ctrl)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				mockService := smsmocks.NewMockService(ctrl)
				mockService.EXPECT().Send(context.Background(), gomock.Any(), gomock.Any()).Return(nil)
				mockService.EXPECT().Send(context.Background(), gomock.Any(), gomock.Any()).Return(nil)
				return mockLimiter, mockService
			},
			after: func(service *TooManyErrFailoverService, err error) error {
				return service.Send(context.Background(), "123", []string{})
			},
			wantIdx: 1,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			limiter, service := tc.mock(ctrl)
			failoverService := NewTooManyErrFailoverService([]sms.Service{service, service, service}, limiter)
			err := failoverService.Send(context.Background(), "123", []string{})
			err = tc.after(failoverService, err)
			assert.Equal(t, tc.wantIdx, failoverService.idx)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
