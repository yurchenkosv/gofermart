package service

import (
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
)

func CreateOrder(order *model.Order, repository *dao.PostgresRepository) error {
	checkOrder, _ := repository.GetOrderByNumber(order.Number)
	if checkOrder.ID != nil {
		if checkOrder.User.ID == order.User.ID {
			return &errors.OrderAlreadyAcceptedCurrentUserError{
				User:        order.User.Login,
				OrderNumber: order.Number,
			}
		} else {
			return &errors.OrderAlreadyAcceptedDifferentUser{
				OrderNumber: order.Number,
			}
		}

	}
	if !checkOrderFormat(order.Number) {
		return &errors.OrderFormatError{
			OrderNumber: order.Number,
		}
	}
	repository.SetOrder(*order).Save()
	return nil
}

func GetUploadedOrdersForUser(order *model.Order, repository *dao.PostgresRepository) ([]model.Order, error) {
	orders, err := repository.GetOrders(*order)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, &errors.NoOrdersDataError{}
	}
	return orders, nil
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
