package ioc

import (
	"webook/internal/service/sms"
	"webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	//服务选择
	return memory.NewService()
}
