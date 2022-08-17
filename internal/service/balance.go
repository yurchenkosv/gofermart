package service

import (
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/model"
)

type Balance interface {
	GetCurrentUserBalance(UserID int) (*model.Balance, error)
}

type BalanceService struct {
	repo dao.Repository
}

func NewBalance(repo dao.Repository) Balance {
	return BalanceService{repo: repo}
}

func (b BalanceService) GetCurrentUserBalance(UserID int) (*model.Balance, error) {
	return b.repo.GetBalanceByUserID(UserID)
}
