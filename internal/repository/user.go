package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

type UserRepository struct {
	dao *dao.UserDAO
}

var (
	ErrUserNotFind = dao.ErrUserNotFind
)

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{dao: dao}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    user.Email,
		Password: user.Password,
	})
}
func (r *UserRepository) FindById(ctx context.Context, user domain.User) error {
	//先从cache找
	//再从dao里面找
	//找到了回写cache
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, user domain.User) (domain.User, error) {
	result, err := r.dao.FindByEmail(ctx, user.Email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Email:    result.Email,
		Password: result.Password,
	}, err
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
