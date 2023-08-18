package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webook/internal/domain"
	"webook/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

var (
	ErrInvalidUserOrPassword = errors.New("邮箱或者密码不对")
	ErrUserNotFind           = repository.ErrUserNotFind
)

func NewUserService(userRepository *repository.UserRepository) *UserService {
	return &UserService{repo: userRepository}
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	//加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	//存储
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, u domain.User) (domain.User, error) {
	result, err := svc.repo.FindByEmail(ctx, u)
	if err != nil {
		return result, err
	}
	if errors.Is(err, ErrUserNotFind) {
		return result, ErrInvalidUserOrPassword
	}

	//判断哈希
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(u.Password))
	if err != nil {
		return result, ErrInvalidUserOrPassword
	}
	return result, err
}

func (svc *UserService) EditUserPassword(ctx context.Context, u domain.User) error {
	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(password)
	err = svc.repo.Update(ctx, u)
	if err != nil {
		return err
	}
	return err
}
