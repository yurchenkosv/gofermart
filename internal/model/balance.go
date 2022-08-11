package model

type Balance struct {
	ID           int     `json:"-"`
	User         User    `json:"-"`
	Balance      float32 `json:"current"`
	SpentAllTime float32 `json:"withdrawn"`
}
