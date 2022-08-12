package service

import (
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
)

func RegisterUser(user *model.User, repository dao.Repository) (*model.User, error) {
	savedUser, _ := repository.GetUser(user)
	if savedUser.ID != nil {
		err := errors.UserAlreadyExistsError{User: user.Login}
		return nil, &err
	}
	err := repository.Save(user)
	if err != nil {
		return nil, err
	}
	savedUser, _ = repository.GetUser(user)
	return savedUser, nil
}

func AuthenticateUser(user *model.User, repository dao.Repository) (*model.User, error) {
	user, _ = repository.GetUser(user)
	if user.ID == nil {
		err := errors.InvalidUserError{}
		return nil, &err
	}
	return user, nil
}
