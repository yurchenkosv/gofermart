package dao

import "github.com/yurchenkosv/gofermart/internal/model"

type Repository interface {
	GetUser() *model.User
}
