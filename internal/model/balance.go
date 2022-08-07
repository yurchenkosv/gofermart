package model

type Balance struct {
	Id           int
	User         User
	Balance      float64 `json:"current"`
	SpentAllTime float64 `json:"withdrawn"`
}
