package model

import (
	"encoding/json"
	"time"
)

type Order struct {
	ID         *int      `json:"-"`
	User       *User     `json:"-"`
	Number     string    `json:"number"`
	Accrual    *float32  `json:"accrual,omitempty"`
	Status     string    `json:"status"`
	UploadTime time.Time `json:"uploaded_at,omitempty"`
}

func (o *Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		Unloaded string `json:"uploaded_at"`
		*Alias
	}{
		Unloaded: o.UploadTime.Format(time.RFC3339),
		Alias:    (*Alias)(o),
	})
}
