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

type BalanceHandler struct {
	balanceService  service.Balance
	withdrawService service.Withdraw
}

func NewBalanceHandler(balanceService *service.Balance, withdrawService *service.Withdraw) BalanceHandler {
	return BalanceHandler{
		balanceService:  *balanceService,
		withdrawService: *withdrawService,
	}
}

func (h BalanceHandler) HandleGetBalance(writer http.ResponseWriter, request *http.Request) {
	userID := GetUserIDFromToken(request.Context())
	balance, err := h.balanceService.GetCurrentUserBalance(userID)
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

func (h BalanceHandler) HandleBalanceWithdraw(writer http.ResponseWriter, request *http.Request) {
	userID := GetUserIDFromToken(request.Context())
	withdraw := model.Withdraw{
		ProcessedAt: time.Now(),
		User:        model.User{ID: &userID},
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Error(err)
		return
	}
	err = json.Unmarshal(body, &withdraw)
	if err != nil {
		log.Error(err)
		return
	}

	err = h.withdrawService.ProcessWithdraw(withdraw)
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

func (h BalanceHandler) HandleGetBalanceWithdraws(writer http.ResponseWriter, request *http.Request) {
	userID := GetUserIDFromToken(request.Context())

	withdrawals, err := h.withdrawService.GetWithdrawalsForCurrentUser(userID)
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
