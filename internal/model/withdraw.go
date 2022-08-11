package model

import "time"

type Withdraw struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
	User        User      `json:"-"`
}
