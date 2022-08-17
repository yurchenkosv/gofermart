package handlers

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"github.com/yurchenkosv/gofermart/internal/service"
	"io"
	"net/http"
)

func HandleUserRegistration(writer http.ResponseWriter, request *http.Request) {
	var user model.User

	data, err := io.ReadAll(request.Body)

	if err != nil {
		log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(data, &user)
	if err != nil {
		log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	cfg := GetConfigFromContext(request.Context())
	auth := service.NewAuthService(cfg.Repo)

	updatedUser, err := auth.RegisterUser(&user)
	if err != nil {
		switch e := err.(type) {
		case *errors.UserAlreadyExistsError:
			log.Error(err)
			writer.WriteHeader(http.StatusConflict)
			return
		default:
			log.Error("error creating user", e)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	writer = SetToken(writer, request, *updatedUser)
	writer.WriteHeader(http.StatusOK)
}

func HandleUserLogin(writer http.ResponseWriter, request *http.Request) {
	var user model.User

	data, err := io.ReadAll(request.Body)
	if err != nil {
		log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, &user)
	if err != nil {
		log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	cfg := GetConfigFromContext(request.Context())
	auth := service.NewAuthService(cfg.Repo)

	updatedUser, err := auth.AuthenticateUser(&user)
	if err != nil {
		switch e := err.(type) {

		case *errors.InvalidUserError:
			log.Error(err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		default:
			log.Error("error during user authentication ", e)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	writer = SetToken(writer, request, *updatedUser)
	writer.WriteHeader(http.StatusOK)
}
