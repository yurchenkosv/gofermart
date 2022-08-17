package handlers

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"github.com/yurchenkosv/gofermart/internal/service"
	"io"
	"net/http"
	"strings"
	"time"
)

type OrdersHanlder struct {
	orderService service.Order
}

func NewOrderHandler(orderService *service.Order) OrdersHanlder {
	return OrdersHanlder{orderService: *orderService}

}

func (h OrdersHanlder) HandleCreateOrder(writer http.ResponseWriter, request *http.Request) {

	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID := GetUserIDFromToken(request.Context())
	orderNum := strings.TrimSpace(string(body))

	order := model.Order{
		User:       &model.User{ID: &userID},
		Number:     orderNum,
		Status:     model.OrderStatusNew,
		UploadTime: time.Now(),
	}

	log.Infof("creating order with number %s, by user %d", orderNum, userID)

	err = h.orderService.CreateOrder(&order)
	if err != nil {
		switch err.(type) {
		case *errors.OrderAlreadyAcceptedDifferentUserError:
			log.Error(err)
			writer.WriteHeader(http.StatusConflict)
			return
		case *errors.OrderAlreadyAcceptedCurrentUserError:
			log.Error(err)
			writer.WriteHeader(http.StatusOK)
			return
		case *errors.OrderFormatError:
			log.Error(err)
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		default:
			log.Error("error creating order", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	writer.WriteHeader(http.StatusAccepted)
}

func (h OrdersHanlder) HandleGetOrders(writer http.ResponseWriter, request *http.Request) {
	userID := GetUserIDFromToken(request.Context())
	log.Infof("getting all orders with user %d", userID)
	orders, err := h.orderService.GetUploadedOrdersForUser(userID)
	if err != nil {
		switch err.(type) {
		case *errors.NoOrdersError:
			log.Error(err)
			writer.WriteHeader(http.StatusNoContent)
			return
		default:
			log.Error("error getting order ", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	result, err := json.Marshal(orders)
	if err != nil {
		log.Error(err)
	}
	writer.Header().Add("Content-Type", "application/json")
	writer.Write(result)
}
