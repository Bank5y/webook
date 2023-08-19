package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/internal/repository"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/internal/web/middleware"
)

func main() {
	db := initDB()
	server := initWebServer()

	u := initUser(db)
	u.RegisterRouter(server)

	server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	//跨域处理
	server.Use(newCors())

	//session处理
	//store := cookie.NewStore([]byte("secret"))
	//store := memstore.NewStore([]byte("tbkykLFqpai8IwdLt9N20HfAsFZoK1uA"), []byte("Gv08GPb5tXIjrtQ8m2cwAVukIkUkDBLG"))
	//size:最大空闲连接数 network:tcp协议 address:链接学习 password:密码
	store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
		[]byte("tbkykLFqpai8IwdLt9N20HfAsFZoK1uA"), []byte("Gv08GPb5tXIjrtQ8m2cwAVukIkUkDBLG"))
	if err != nil {
		panic(err)
	}

	server.Use(sessions.Sessions("mySessions", store))

	//验证登录状态
	server.Use(
		middleware.NewLoginMiddlewareBuilder().
			Ignore("/users/login").
			Ignore("/users/signup").
			Build())

	return server
}

// 跨域中间件
func newCors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"https://localhost"},
		AllowCredentials: true,
	})
}

// 初始化数据库
func initDB() *gorm.DB {

	//尝试链接数据库
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		//panic相当于整个goroutine结束
		//整个goroutine结束
		panic(err)
	}
	//建表初始化
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

// 初始化UserHandler
func initUser(db *gorm.DB) *web.UserHandler {

	ud := dao.NewUserDao(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}
