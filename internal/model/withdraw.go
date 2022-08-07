package model

import "time"

type Withdraw struct {
	Order       Order
	Sum         int
	ProcessedAt time.Time
	User        User
}
