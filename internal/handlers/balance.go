package handlers

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"github.com/yurchenkosv/gofermart/internal/service"
	"io"
	"net/http"
	"time"
)

func HandleGetBalance(writer http.ResponseWriter, request *http.Request) {
	userID := GetUserIDFromToken(request.Context())
	cfg := GetConfigFromContext(request.Context())
	balanceService := service.NewBalance(cfg.Repo)

	b := model.Balance{User: model.User{ID: &userID}}
	balance, err := balanceService.GetCurrentUserBalance(b)

	if err != nil {
		log.Error("error getting balance", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(balance)
	if err != nil {
		log.Error("error marshalling to json", err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
	writer.Header().Add("Content-Type", "application/json")
	writer.Write(body)
}

func HandleBalanceWithdraw(writer http.ResponseWriter, request *http.Request) {
	userID := GetUserIDFromToken(request.Context())
	cfg := GetConfigFromContext(request.Context())
	withdraw := model.Withdraw{User: model.User{ID: &userID}}
	withdrawService := service.NewWithdrawService(cfg.Repo)

	body, _ := io.ReadAll(request.Body)
	err := json.Unmarshal(body, &withdraw)
	if err != nil {
		log.Error(err)
	}

	withdraw.ProcessedAt = time.Now()

	err = withdrawService.ProcessWithdraw(withdraw)
	if err != nil {
		switch err.(type) {
		case *errors.LowBalanceError:
			log.Error(err)
			writer.WriteHeader(http.StatusPaymentRequired)
			return
		case *errors.OrderFormatError:
			log.Error(err)
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		default:
			log.Error("error process withdraw", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func HandleGetBalanceWithdraws(writer http.ResponseWriter, request *http.Request) {
	userID := GetUserIDFromToken(request.Context())
	cfg := GetConfigFromContext(request.Context())
	withdraw := model.Withdraw{User: model.User{ID: &userID}}
	withdrawService := service.NewWithdrawService(cfg.Repo)

	withdrawals, err := withdrawService.GetWithdrawalsForCurrentUser(withdraw)
	if err != nil {
		switch err.(type) {
		case *errors.NoWithdrawalsError:
			log.Error(err)
			writer.WriteHeader(http.StatusNoContent)
		default:
			log.Error("error getting withdrawals", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	body, err := json.Marshal(withdrawals)
	if err != nil {
		log.Error("error marshalling to json", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Header().Add("Content-Type", "application/json")
	writer.Write(body)
}
