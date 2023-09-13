package web

type Result struct {
	//业务错误码
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
