package services

import (
	"errors"
	"imooc-product/datamodels"
	"imooc-product/repositories"

	"golang.org/x/crypto/bcrypt"

	"github.com/kataras/golog"
)

// 用户业务逻辑：面向用户
type IUserService interface {
	// 判断用户提供的用户名和密码是否匹配
	IsPwdSussess(userName, password string) (*datamodels.User, bool)
	// 用户注册
	AddUser(*datamodels.User) (int64, error)
	GetUserByID(userID int64) (*datamodels.User, error)
	GetUserByName(userName string) (*datamodels.User, error)
}

type UserService struct {
	UserRepository repositories.IUserRepository
}

func NewUserService(userRepository repositories.IUserRepository) IUserService {
	return &UserService{userRepository}
}

func (u *UserService) IsPwdSussess(userName, password string) (*datamodels.User, bool) {
	user, err := u.UserRepository.Select(userName)
	if err != nil {
		golog.Debug(err)
		return nil, false
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(password))
	if err != nil {
		golog.Debug(err)
		return nil, false
	}
	return user, true
}
func (u *UserService) AddUser(user *datamodels.User) (int64, error) {
	// 传入的user结构体为用户提交的，其中密码部分需要先加密
	pwd, err := bcrypt.GenerateFromPassword([]byte(user.HashPassword), bcrypt.DefaultCost)
	if err != nil {
		return -1, err
	}
	user.HashPassword = string(pwd)
	return u.UserRepository.Insert(user)
}
func (u *UserService) GetUserByName(userName string) (*datamodels.User, error) {
	if userName == "" {
		return nil, errors.New("传入数据不正确")
	}
	user, err := u.UserRepository.Select(userName)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (u *UserService) GetUserByID(userID int64) (*datamodels.User, error) {
	user, err := u.UserRepository.SelectByID(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
