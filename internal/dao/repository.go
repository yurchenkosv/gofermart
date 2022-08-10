package dao

import "github.com/yurchenkosv/gofermart/internal/model"

type Repository interface {
	GetWithdrawals(withdraw model.Withdraw) ([]*model.Withdraw, error)
}
