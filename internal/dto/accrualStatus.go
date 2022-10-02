package dto

type AccrualStatus struct {
	OrderNum string   `json:"order"`
	Status   string   `json:"status"`
	Accrual  *float32 `json:"accrual,omitempty"`
}
