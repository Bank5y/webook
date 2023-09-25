package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"webook/internal/service/sms"
)

type SMSService struct {
	svc sms.Service
	key string
}

func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	var tc TokenClaims

	//如果成功解析,说明就是对应的业务方
	//没有error就说明,token是我发的
	token, err := jwt.ParseWithClaims(biz, &tc, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token不合法")
	}

	return s.svc.Send(ctx, biz, args, numbers...)
}

type TokenClaims struct {
	jwt.RegisteredClaims
	Tpl string
}
