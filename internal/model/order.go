package model

import "time"

type Order struct {
	ID         *int      `json:"-"`
	User       *User     `json:"-"`
	Number     string    `json:"number"`
	Accrual    *float32  `json:"accrual,omitempty"`
	Status     string    `json:"status"`
	UploadTime time.Time `json:"uploaded_at,omitempty"`
}
