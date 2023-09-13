package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/config"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/service/sms/memory"
	"webook/internal/web"
	"webook/internal/web/middleware"
)

func main() {
	//初始化数据库
	db := initDB()
	//初始化Redis缓存
	rdb := initRedis()

	//初始化http服务器
	server := initWebServer()

	//初始化注册handler
	u := initUser(db, rdb)
	u.RegisterRouter(server)

	server.Run(":8186")
}

// 初始化中间件
func initWebServer() *gin.Engine {
	server := gin.Default()

	////限流处理
	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: config.Config.Redis.Addr,
	//})
	//server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	//跨域处理
	server.Use(newCors())

	//session处理
	//store := cookie.NewStore([]byte("secret"))
	store := memstore.NewStore([]byte("tbkykLFqpai8IwdLt9N20HfAsFZoK1uA"), []byte("Gv08GPb5tXIjrtQ8m2cwAVukIkUkDBLG"))
	//size:最大空闲连接数 network:tcp协议 address:链接学习 password:密码
	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
	//	[]byte("tbkykLFqpai8IwdLt9N20HfAsFZoK1uA"), []byte("Gv08GPb5tXIjrtQ8m2cwAVukIkUkDBLG"))
	//if err != nil {
	//	panic(err)
	//}

	server.Use(sessions.Sessions("mySessions", store))

	//验证登录状态
	//server.Use(middleware.NewLoginMiddlewareBuilder().
	//	Ignore("/users/login").
	//	Ignore("/users/signup").
	//	Build())
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		Ignore("/users/login").
		Ignore("/users/login_sms/code/send").
		Ignore("/users/login_sms").
		Ignore("/users/signup").Build())

	return server
}

// 跨域中间件
func newCors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"https://localhost"},
		//传递信息跨域
		ExposeHeaders:    []string{"jwt-token"},
		AllowHeaders:     []string{"Authorization"},
		AllowCredentials: true,
	})
}

// 初始化数据库
func initDB() *gorm.DB {

	//尝试链接数据库
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
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

// 初始化Redis缓存
func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	return redisClient
}

// 初始化UserHandler
func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {

	ud := dao.NewUserDao(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(ud, uc)
	svc := service.NewUserService(repo)
	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	cs := service.NewCodeService(smsSvc, codeRepo)
	u := web.NewUserHandler(svc, cs)
	return u
}
