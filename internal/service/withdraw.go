package service

import (
	"errors"
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/model"
)

func GetWithdrawalsForCurrentUser(withdraw *model.Withdraw, repo dao.PostgresRepository) ([]*model.Withdraw, error) {
	return nil, errors.New("")
}
