package model

type User struct {
	Id       *int
	Login    string `json:"login"`
	Password string `json:"password"`
}
