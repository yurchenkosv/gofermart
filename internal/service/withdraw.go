package service

import (
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"strconv"
)

func GetWithdrawalsForCurrentUser(withdraw model.Withdraw, repo dao.Repository) ([]*model.Withdraw, error) {
	withdrawals, err := repo.GetWithdrawals(withdraw)
	if err != nil {
		return nil, err
	}
	if len(withdrawals) == 0 {
		return nil, &errors.NoWithdrawalsError{}
	}
	return withdrawals, nil
}

func ProcessWithdraw(withdraw model.Withdraw, repository dao.Repository) error {
	orderNum, _ := strconv.Atoi(withdraw.Order)
	if !checkOrderFormat(orderNum) {
		return &errors.OrderFormatError{OrderNumber: withdraw.Order}
	}
	b := model.Balance{User: model.User{ID: withdraw.User.ID}}
	currentBalance, _ := repository.GetBalance(b)

	expectedAfterWithdraw := currentBalance.Balance - withdraw.Sum
	if expectedAfterWithdraw < 0 {
		return errors.LowBalanceError{
			CurrentBalance: currentBalance.Balance,
		}
	}
	b.Balance = expectedAfterWithdraw
	b.SpentAllTime = currentBalance.SpentAllTime + withdraw.Sum

	//TODO надо делать списание и обновление баланса в одной транзакции
	repository.Save(&b)
	repository.Save(&withdraw)
	return nil
}
