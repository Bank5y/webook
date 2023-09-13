package repository

import (
	"context"
	"database/sql"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

var (
	ErrUserNotFind = dao.ErrUserNotFind
)

func NewUserRepository(dao *dao.UserDAO, userCache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: userCache,
	}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(user))
}

func (r *UserRepository) FindById(ctx context.Context, user domain.User) (domain.User, error) {
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

func (r *UserRepository) FindByEmail(ctx context.Context, user domain.User) (domain.User, error) {
	result, err := r.dao.FindByEmail(ctx, user.Email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(result), err
}

func (r *UserRepository) FindByPhone(ctx context.Context, user domain.User) (domain.User, error) {
	result, err := r.dao.FindByPhone(ctx, user.Phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(result), err
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) error {
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

func (r *UserRepository) entityToDomain(ud dao.User) domain.User {
	return domain.User{
		Id:       ud.Id,
		Email:    ud.Email.String,
		Phone:    ud.Phone.String,
		Password: ud.Password,
	}
}

func (r *UserRepository) domainToEntity(ud domain.User) dao.User {
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
