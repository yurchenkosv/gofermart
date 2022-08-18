package clients

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/dto"
)

type AccrualProvider interface {
	GetOrderStatusByOrderNum(orderNum int) (*dto.AccrualStatus, error)
}

type AccrualClient struct {
	accruaSysAddress string
}

func NewAccrualClient(accrualAddress string) *AccrualClient {
	return &AccrualClient{accruaSysAddress: accrualAddress}
}

func (c AccrualClient) GetOrderStatusByOrderNum(orderNum int) (*dto.AccrualStatus, error) {
	var (
		accrualStatus = dto.AccrualStatus{}
	)
	client := resty.New().
		SetBaseURL(c.accruaSysAddress).
		SetRetryCount(3)
	resp, err := client.R().
		Get(fmt.Sprintf("/api/orders/%d", orderNum))
	if err != nil {
		log.Error("error sending request to accrual system", err)
		return nil, err
	}
	log.Info("received responce from accrual system: ", string(resp.Body()))
	err = json.Unmarshal(resp.Body(), &accrualStatus)
	if err != nil {
		log.Error("error unmarshalling json: ", err)
		return nil, err
	}
	return &accrualStatus, nil
}
