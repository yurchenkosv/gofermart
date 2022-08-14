package service

import (
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/model"
)

type Balance interface {
	GetCurrentUserBalance(balance model.Balance) (*model.Balance, error)
}

type BalanceService struct {
	repo dao.Repository
}

func NewBalance(repo dao.Repository) BalanceService {
	return BalanceService{repo: repo}
}

func (b BalanceService) GetCurrentUserBalance(balance model.Balance) (*model.Balance, error) {
	return b.repo.GetBalance(balance)
}
