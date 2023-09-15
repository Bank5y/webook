package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/config"
	"webook/internal/repository/dao"
)

func InitDB() *gorm.DB {
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
