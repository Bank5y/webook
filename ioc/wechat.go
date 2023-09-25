package ioc

import (
	"webook/internal/service/oauth2/wechat"
)

func InitWechatService() *wechat.DevService {
	//appId, ok := os.LookupEnv("WECHAT_APP_ID")
	//if !ok {
	//	panic("没有找到环境变量 WECHAT_APP_ID ")
	//}
	//appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	//if !ok {
	//	panic("没有找到环境变量 WECHAT_APP_SECRET")
	//}

	appId := ""
	appKey := ""
	return wechat.NewDevService(appId, appKey)
}
