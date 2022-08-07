package handlers

import "net/http"

func HandleGetBalance(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("get balance placeholder"))
	writer.WriteHeader(http.StatusOK)
}

func HandleBalanceDraw(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("balance draw placeholder"))
	writer.WriteHeader(http.StatusOK)
}

func HandleGetBalanceDraws(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("balance get draws placeholder"))
	writer.WriteHeader(http.StatusOK)
}
