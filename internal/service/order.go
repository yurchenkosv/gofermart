package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"strconv"
	"sync"
)

var (
	mux sync.Mutex
)

type Order interface {
	CreateOrder(order *model.Order) error
	GetUploadedOrdersForUser(order *model.Order) ([]model.Order, error)
	UpdateOrderStatus(order model.Order) error
}

type OrderService struct {
	repo dao.Repository
}

func NewOrderService(repo dao.Repository) OrderService {
	return OrderService{repo: repo}
}

func (s OrderService) CreateOrder(order *model.Order) error {
	checkOrder, err := s.repo.GetOrderByNumber(order.Number)
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
	err = s.repo.Save(order)
	if err != nil {
		return err
	}
	return nil
}

func (s OrderService) GetUploadedOrdersForUser(order *model.Order) ([]model.Order, error) {
	orders, err := s.repo.GetOrdersForUser(*order)
	log.Info("found orders for current user: ", orders)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, &errors.NoOrdersError{}
	}
	return orders, nil
}

func (s OrderService) UpdateOrderStatus(order model.Order) error {
	mux.Lock()
	defer mux.Unlock()
	orderInDB, err := s.repo.GetOrderByNumber(order.Number)
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

	err = s.repo.Save(orderInDB)
	if err != nil {
		return err
	}

	if order.Accrual != nil {
		balance, err := s.repo.GetBalance(model.Balance{
			User: model.User{
				ID: orderInDB.User.ID,
			},
		})
		if err != nil {
			log.Error(err)
			return err
		}
		balance.Balance += *orderInDB.Accrual
		err = s.repo.Save(balance)
		if err != nil {
			return err
		}
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
