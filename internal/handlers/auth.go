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

	CheckErrors(err, writer)
	err = json.Unmarshal(data, &user)
	CheckErrors(err, writer)
	cfg := GetConfigFromContext(request.Context())
	repo := cfg.Repo

	updatedUser, err := service.RegisterUser(&user, repo)
	if err != nil {
		switch e := err.(type) {
		case *errors.UserAlreadyExistsError:
			log.Error(err)
			writer.WriteHeader(http.StatusConflict)
		default:
			log.Error("error creating user", e)
			CheckErrors(e, writer)
		}
	}
	writer = *SetToken(writer, request, *updatedUser)
	writer.WriteHeader(http.StatusOK)
}

func HandleUserLogin(writer http.ResponseWriter, request *http.Request) {
	var user model.User

	data, err := io.ReadAll(request.Body)
	CheckErrors(err, writer)

	err = json.Unmarshal(data, &user)
	CheckErrors(err, writer)

	cfg := GetConfigFromContext(request.Context())
	repo := cfg.Repo

	updatedUser, err := service.AuthenticateUser(&user, repo)
	if err != nil {
		switch e := err.(type) {

		case *errors.InvalidUserError:
			log.Error(err)
			writer.WriteHeader(http.StatusUnauthorized)
		default:
			log.Error("error during user authentication ", e)
			CheckErrors(e, writer)
		}
	}
	writer = *SetToken(writer, request, *updatedUser)
	writer.WriteHeader(http.StatusOK)
}
