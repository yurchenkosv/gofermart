package model

import (
	"encoding/json"
	"time"
)

type Withdraw struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
	User        User      `json:"-"`
}

func (w *Withdraw) MarshalJSON() ([]byte, error) {
	type Alias Withdraw
	return json.Marshal(&struct {
		Unloaded string `json:"uploaded_at"`
		*Alias
	}{
		Unloaded: w.ProcessedAt.Format(time.RFC3339),
		Alias:    (*Alias)(w),
	})
}
