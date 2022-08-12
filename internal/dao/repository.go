package dao

import "github.com/yurchenkosv/gofermart/internal/model"

type Repository interface {
	GetUser(user *model.User) (*model.User, error)
	GetOrderByNumber(orderNumber string) (*model.Order, error)
	GetOrdersForStatusUpdate() ([]*model.Order, error)
	GetOrdersForUser(order model.Order) ([]model.Order, error)
	GetBalance(balance model.Balance) (*model.Balance, error)
	GetWithdrawals(withdraw model.Withdraw) ([]*model.Withdraw, error)
	Save(obj interface{}) error
}
