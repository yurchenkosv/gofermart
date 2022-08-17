package dao

import "github.com/yurchenkosv/gofermart/internal/model"

type Repository interface {
	GetUser(user *model.User) (*model.User, error)
	GetOrderByNumber(orderNumber string) (*model.Order, error)
	GetOrdersForStatusUpdate() ([]*model.Order, error)
	GetOrdersByUserID(userID int) ([]model.Order, error)
	GetBalanceByUserID(userID int) (*model.Balance, error)
	GetWithdrawalsByUserID(userID int) ([]*model.Withdraw, error)
	SaveWithdraw(withdraw *model.Withdraw) error
	SaveBalance(balance *model.Balance) error
	SaveOrder(order *model.Order) error
	SaveUser(user *model.User) error
	Shutdown()
}
