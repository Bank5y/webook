package dao

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name string
		mock func(t *testing.T) *sql.DB

		ctx  context.Context
		user User

		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				res := sqlmock.NewResult(3, 1)
				mock.ExpectExec("INSERT INTO `users`").WillReturnResult(res)
				require.NoError(t, err)
				return mockDB
			},
			ctx: context.Background(),
			user: User{
				Email: sql.NullString{
					String: "1234@qq.com",
					Valid:  true,
				},
			},
			wantErr: nil,
		},
		{
			name: "插入失败-邮箱冲突",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(&mysql.MySQLError{
					Number: 1062,
				})
				require.NoError(t, err)
				return mockDB
			},
			ctx:     context.Background(),
			user:    User{},
			wantErr: ErrUserDuplicateEmail,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn: tc.mock(t),
				//跳过调用show version
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				//不ping数据库
				DisableAutomaticPing: true,
				//跳过开启事务
				SkipDefaultTransaction: true,
			})
			d := NewUserDao(db)
			u := tc.user
			err = d.Insert(tc.ctx, u)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
