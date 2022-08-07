package model

import "time"

type Order struct {
	ID         *int      `json:"-"`
	User       *User     `json:"-"`
	Number     int       `json:"number"`
	Accrual    *int      `json:"accrual,omitempty"`
	Status     string    `json:"status"`
	UploadTime time.Time `json:"uploaded_at"`
}
