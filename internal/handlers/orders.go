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

func HandleCreateOrder(writer http.ResponseWriter, request *http.Request) {
	var order model.Order
	cfg := GetConfigFromContext(request.Context())
	repo := cfg.Repo
	now := time.Now()

	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	orderNum := strings.TrimSpace(string(body))
	order.Number = orderNum

	userID := GetUserIDFromToken(request.Context())
	order.Status = model.OrderStatusNew
	order.User = &model.User{ID: &userID}
	order.UploadTime = now
	orderService := service.NewOrderService(repo)

	log.Infof("creating order with number %s, by user %d", orderNum, userID)

	err = orderService.CreateOrder(&order)
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

func HandleGetOrders(writer http.ResponseWriter, request *http.Request) {
	userID := GetUserIDFromToken(request.Context())
	cfg := GetConfigFromContext(request.Context())
	repo := cfg.Repo
	order := model.Order{
		User: &model.User{
			ID: &userID,
		},
	}
	log.Infof("getting all orders with user %d", userID)
	orderSerivice := service.NewOrderService(repo)
	orders, err := orderSerivice.GetUploadedOrdersForUser(&order)
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
