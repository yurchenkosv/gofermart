package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/config"
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/dto"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"github.com/yurchenkosv/gofermart/internal/service"
	"strconv"
)

func UpdateOrderStatusFromAccrualSys(order int, config *config.ServerConfig, repo dao.Repository) {
	var (
		accrualStatus = dto.AccrualStatus{}
	)
	client := resty.New().
		SetBaseURL(config.AccrualSystemAddress).
		SetRetryCount(3)
	resp, err := client.R().
		Get(fmt.Sprintf("/api/orders/%d", order))
	if err != nil {
		log.Error("error sending request to accrual system", err)
		return
	}
	log.Info("received responce from accrual system: ", string(resp.Body()))
	err = json.Unmarshal(resp.Body(), &accrualStatus)
	if err != nil {
		log.Error("error unmarshalling json: ", err)
		return
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
			return
		case *errors.OrderNoChangeError:
			log.Warnf("order %s status not updated yet %s", orderToUpdate.Number, err)
		default:
			log.Error("error updating order: ", err)
			return
		}
	}
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
			UpdateOrderStatusFromAccrualSys(orderNum, cfg, repo)
		}()
	}
}
