package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFind        = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (u *UserDAO) Insert(ctx context.Context, user User) error {
	//更新Ctime Utime
	now := time.Now().UnixMilli()
	user.Ctime = now
	user.Utime = now
	var mysqlErr *mysql.MySQLError
	err := u.db.WithContext(ctx).Create(&user).Error
	if errors.As(err, &mysqlErr) {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			//唯一索引邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (u *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var result User
	err := u.db.WithContext(ctx).Where("email=?", email).First(&result).Error
	return result, err
}

func (u *UserDAO) Update(ctx context.Context, user User) error {
	//更新Ctime Utime
	now := time.Now().UnixMilli()
	user.Utime = now

	//更新
	return u.db.WithContext(ctx).Where("email=?", user.Email).Save(&user).Error
}

// User 直接对应数据库表
type User struct {
	id       int64  `gorm:"primaryKey,autoIncrement"`
	Email    string `gorm:"unique"`
	Password string

	//Create time ms
	Ctime int64
	//update time ms
	Utime int64
}
