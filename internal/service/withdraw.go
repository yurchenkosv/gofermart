package service

import (
	"context"
	"database/sql"
	errors2 "errors"
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"strconv"
)

type Withdraw interface {
	GetWithdrawalsForCurrentUser(UserID int) ([]*model.Withdraw, error)
	ProcessWithdraw(withdraw model.Withdraw) error
}

type WithdrawService struct {
	repo dao.Repository
}

func NewWithdrawService(repo dao.Repository) Withdraw {
	return WithdrawService{repo: repo}
}

func (s WithdrawService) GetWithdrawalsForCurrentUser(UserID int) ([]*model.Withdraw, error) {
	withdrawals, err := s.repo.GetWithdrawalsByUserID(UserID)
	if err != nil {
		return nil, err
	}
	if len(withdrawals) == 0 {
		return nil, &errors.NoWithdrawalsError{}
	}
	return withdrawals, nil
}

func (s WithdrawService) ProcessWithdraw(withdraw model.Withdraw) error {
	orderNum, _ := strconv.Atoi(withdraw.Order)
	if !checkOrderFormat(orderNum) {
		return &errors.OrderFormatError{OrderNumber: withdraw.Order}
	}
	b := model.Balance{User: model.User{ID: withdraw.User.ID}}
	currentBalance, err := s.repo.GetBalanceByUserID(*b.User.ID)
	if errors2.Is(err, sql.ErrNoRows) {
		return &errors.LowBalanceError{}
	}

	expectedAfterWithdraw := currentBalance.Balance - withdraw.Sum
	if expectedAfterWithdraw < 0 {
		return &errors.LowBalanceError{
			CurrentBalance: currentBalance.Balance,
		}
	}
	b.Balance = expectedAfterWithdraw
	b.SpentAllTime = currentBalance.SpentAllTime + withdraw.Sum
	ctx := context.Background()
	err = s.repo.Atomic(ctx, func(r dao.Repository) error {
		err = s.repo.SaveBalance(&b)
		if err != nil {
			return err
		}
		err = s.repo.SaveWithdraw(&withdraw)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
