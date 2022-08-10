package model

type Balance struct {
	ID           int  `json:"-"`
	User         User `json:"-"`
	Balance      int  `json:"current"`
	SpentAllTime int  `json:"withdrawn"`
}
