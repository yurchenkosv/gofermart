package service

import (
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/model"
)

func GetCurrentUserBalance(balance model.Balance, repository dao.Repository) (*model.Balance, error) {
	return repository.GetBalance(balance)
}
