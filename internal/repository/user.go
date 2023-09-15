package repository

import (
	"context"
	"database/sql"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindById(ctx context.Context, user domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, user domain.User) (domain.User, error)
	FindByPhone(ctx context.Context, user domain.User) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
}

type UserCacheRepository struct {
	dao   *dao.UserDAO
	cache cache.UserCache
}

var (
	ErrUserNotFind = dao.ErrUserNotFind
)

func NewUserCacheRepository(dao *dao.UserDAO, userCache cache.UserCache) UserRepository {
	return &UserCacheRepository{
		dao:   dao,
		cache: userCache,
	}
}

func (r *UserCacheRepository) Create(ctx context.Context, user domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(user))
}

func (r *UserCacheRepository) FindById(ctx context.Context, user domain.User) (domain.User, error) {
	//先从cache找
	u, err := r.cache.Get(ctx, user.Id)
	if err == nil {
		return u, err
	}
	//再从dao里面找
	userDb, err := r.dao.FindById(ctx, u.Id)
	if err != nil {
		return domain.User{}, err
	}
	u = r.entityToDomain(userDb)

	//找到了回写cache
	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			//打日志,做监控
		}
	}()

	return domain.User{}, nil
}

func (r *UserCacheRepository) FindByEmail(ctx context.Context, user domain.User) (domain.User, error) {
	result, err := r.dao.FindByEmail(ctx, user.Email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(result), err
}

func (r *UserCacheRepository) FindByPhone(ctx context.Context, user domain.User) (domain.User, error) {
	result, err := r.dao.FindByPhone(ctx, user.Phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(result), err
}

func (r *UserCacheRepository) Update(ctx context.Context, user domain.User) error {
	result, err := r.dao.FindByEmail(ctx, user.Email)
	if err != nil {
		return err
	}
	result.Password = user.Password
	err = r.dao.Update(ctx, result)
	if err != nil {
		return err
	}
	return err
}

func (r *UserCacheRepository) entityToDomain(ud dao.User) domain.User {
	return domain.User{
		Id:       ud.Id,
		Email:    ud.Email.String,
		Phone:    ud.Phone.String,
		Password: ud.Password,
	}
}

func (r *UserCacheRepository) domainToEntity(ud domain.User) dao.User {
	return dao.User{
		Id: ud.Id,
		Email: sql.NullString{
			String: ud.Email,
			Valid:  ud.Email != "",
		},
		Phone: sql.NullString{
			String: ud.Phone,
			Valid:  ud.Phone != "",
		},
		Password: ud.Password,
	}
}
