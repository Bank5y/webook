package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webook/internal/domain"
	"webook/internal/repository"
)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, u domain.User) (domain.User, error)
	FindOrCreate(ctx context.Context, u domain.User) (domain.User, error)
	EditUserPassword(ctx context.Context, u domain.User) error
	Profile(ctx context.Context, u domain.User) (domain.User, error)
}

type UserDevService struct {
	repo repository.UserRepository
}

var (
	ErrInvalidUserOrPassword = errors.New("邮箱或者密码不对")
	ErrUserNotFind           = repository.ErrUserNotFind
)

func NewUserDevService(userRepository repository.UserRepository) UserService {
	return &UserDevService{repo: userRepository}
}

func (svc *UserDevService) SignUp(ctx context.Context, u domain.User) error {
	//加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	//存储
	return svc.repo.Create(ctx, u)
}

func (svc *UserDevService) Login(ctx context.Context, u domain.User) (domain.User, error) {
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

func (svc *UserDevService) FindOrCreate(ctx context.Context, u domain.User) (domain.User, error) {
	//快路径
	uResult, err := svc.repo.FindByPhone(ctx, u)
	if err != ErrUserNotFind {
		return uResult, err
	}
	err = svc.repo.Create(ctx, domain.User{
		Phone: u.Phone,
	})
	if err != nil {
		return domain.User{}, err
	}
	user, err := svc.repo.FindByPhone(ctx, u)
	return user, err
}

func (svc *UserDevService) EditUserPassword(ctx context.Context, u domain.User) error {
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

func (svc *UserDevService) Profile(ctx context.Context, u domain.User) (domain.User, error) {
	user, err := svc.repo.FindById(ctx, u)
	return user, err
}
