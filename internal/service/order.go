package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"strconv"
)

func CreateOrder(order *model.Order, repository *dao.PostgresRepository) error {
	checkOrder, err := repository.GetOrderByNumber(order.Number)
	if err != nil {
		return err
	}
	if checkOrder.ID != nil {
		checkUserID := *checkOrder.User.ID
		orderUserID := *order.User.ID
		if checkUserID == orderUserID {
			return &errors.OrderAlreadyAcceptedCurrentUserError{
				UserID:      checkUserID,
				OrderNumber: order.Number,
			}
		} else {
			return &errors.OrderAlreadyAcceptedDifferentUserError{
				OrderNumber: order.Number,
				UserID:      checkUserID,
			}
		}

	}
	orderNum, _ := strconv.Atoi(order.Number)
	if !checkOrderFormat(orderNum) {
		return &errors.OrderFormatError{
			OrderNumber: order.Number,
		}
	}
	repository.SetOrder(order).Save()
	return nil
}

func GetUploadedOrdersForUser(order *model.Order, repository *dao.PostgresRepository) ([]model.Order, error) {
	orders, err := repository.GetOrdersForUser(*order)
	log.Infof("found orders for current user: ", orders)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, &errors.NoOrdersError{}
	}
	return orders, nil
}

func UpdateOrderStatus(order model.Order, repository *dao.PostgresRepository) error {
	orderInDB, err := repository.GetOrderByNumber(order.Number)
	if err != nil {
		return err
	}
	if orderInDB.ID == nil {
		return &errors.NoOrdersError{}
	}
	if orderInDB.Status == order.Status {
		return &errors.OrderNoChangeError{}
	}
	orderInDB.Accrual = order.Accrual
	orderInDB.Status = order.Status

	repository.SetOrder(&order).Save()
	if order.Accrual != nil {
		balance, _ := repository.GetBalance(model.Balance{
			User: model.User{
				ID: orderInDB.User.ID,
			},
		})
		balance.Balance += *order.Accrual
		repository.SetBalance(*balance).Save()
	}
	return nil
}

func checkOrderFormat(number int) bool {
	return (number%10+luhnChecksum(number/10))%10 == 0
}

func luhnChecksum(number int) int {
	var luhn int
	for i := 0; number > 0; i++ {
		cur := number % 10
		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
