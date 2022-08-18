package controllers

import (
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/clients"
	"github.com/yurchenkosv/gofermart/internal/config"
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"github.com/yurchenkosv/gofermart/internal/service"
	"strconv"
)

func UpdateOrderStatusFromAccrualSys(order int, repo dao.Repository, client clients.AccrualProvider) error {
	accrualStatus, err := client.GetOrderStatusByOrderNum(order)
	if err != nil {
		log.Error(err)
		return err
	}

	orderToUpdate := model.Order{
		Number:  accrualStatus.OrderNum,
		Accrual: accrualStatus.Accrual,
		Status:  accrualStatus.Status,
	}

	orderService := service.NewOrderService(repo)

	err = orderService.UpdateOrderStatus(orderToUpdate)
	if err != nil {
		switch err.(type) {
		case *errors.NoOrdersError:
			log.Errorf("no orders found by number %s, %s", orderToUpdate.Number, err)
			return err
		case *errors.OrderNoChangeError:
			log.Warnf("order %s status not updated yet %s", orderToUpdate.Number, err)
		default:
			log.Error("error updating order: ", err)
			return err
		}
	}
	return nil
}

func GetOrdersForStatusCheck(repository dao.Repository) []*model.Order {
	orders, err := repository.GetOrdersForStatusUpdate()
	if err != nil {
		log.Error("error getting orders", err)
	}
	return orders
}

func StatusCheckLoop(cfg *config.ServerConfig, repo dao.Repository) {
	orders := GetOrdersForStatusCheck(repo)
	for i := range orders {
		orderNum, err := strconv.Atoi(orders[i].Number)
		if err != nil {
			log.Error("cannot parse order num", err)
			continue
		}
		go func() {
			accrualClient := clients.NewAccrualClient(cfg.AccrualSystemAddress)
			err = UpdateOrderStatusFromAccrualSys(orderNum, repo, accrualClient)
			if err != nil {
				return
			}
		}()
	}
}
