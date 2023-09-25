package domain

// User DDD中的entity
type User struct {
	Id       int
	Email    string
	Phone    string
	Password string

	WechatInfo WechatInfo
}
