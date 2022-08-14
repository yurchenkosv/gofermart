package service

import (
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
)

type Auth interface {
	RegisterUser(user *model.User) (*model.User, error)
	AuthenticateUser(user *model.User) (*model.User, error)
}

type UserAuth struct {
	repo dao.Repository
}

func NewAuthService(repo dao.Repository) UserAuth {
	return UserAuth{repo: repo}
}

func (auth UserAuth) RegisterUser(user *model.User) (*model.User, error) {
	savedUser, _ := auth.repo.GetUser(user)
	if savedUser.ID != nil {
		err := errors.UserAlreadyExistsError{User: user.Login}
		return nil, &err
	}
	err := auth.repo.Save(user)
	if err != nil {
		return nil, err
	}
	savedUser, _ = auth.repo.GetUser(user)
	return savedUser, nil
}

func (auth UserAuth) AuthenticateUser(user *model.User) (*model.User, error) {
	user, _ = auth.repo.GetUser(user)
	if user.ID == nil {
		err := errors.InvalidUserError{}
		return nil, &err
	}
	return user, nil
}
