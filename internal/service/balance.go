package service

import (
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/model"
)

func GetCurrentUserBalance(balance model.Balance, repository *dao.PostgresRepository) (*model.Balance, error) {
	return repository.GetBalance(balance)
}

func UpdateBalance(balance model.Balance, repository *dao.PostgresRepository) error {
	return repository.Save(&balance)
}
