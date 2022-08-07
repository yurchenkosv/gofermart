package handlers

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"github.com/yurchenkosv/gofermart/internal/service"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HandleCreateOrder(writer http.ResponseWriter, request *http.Request) {
	var order model.Order
	cfg := GetConfigFromContext(request.Context())
	repo := cfg.Repo
	now := time.Now()

	body, err := io.ReadAll(request.Body)
	CheckErrors(err, writer)

	orderNum := strings.TrimSpace(string(body))
	order.Number, err = strconv.Atoi(orderNum)
	CheckErrors(err, writer)

	userID := GetUserIDFromToken(request.Context())
	order.Status = model.OrderStatusNew
	order.User = &model.User{ID: &userID}
	order.UploadTime = now

	err = service.CreateOrder(&order, repo)
	if err != nil {
		switch e := err.(type) {
		case *errors.OrderAlreadyAcceptedDifferentUser:
			log.Error(err)
			writer.WriteHeader(http.StatusConflict)
		case *errors.OrderAlreadyAcceptedCurrentUserError:
			log.Error(err)
			writer.WriteHeader(http.StatusOK)
		case *errors.OrderFormatError:
			log.Error(err)
			writer.WriteHeader(http.StatusUnprocessableEntity)
		default:
			log.Error("error creating order", e)
			CheckErrors(e, writer)
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
	orders, err := service.GetUploadedOrdersForUser(&order, repo)
	switch e := err.(type) {
	case *errors.NoOrdersDataError:
		log.Error(err)
		writer.WriteHeader(http.StatusNoContent)
	default:
		log.Error("error getting order ", e)
		CheckErrors(e, writer)
	}
	result, _ := json.Marshal(orders)
	writer.Header().Add("Content-Type", "application/json")
	writer.Write(result)
}
