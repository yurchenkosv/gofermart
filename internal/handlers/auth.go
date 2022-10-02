package handlers

import (
	"encoding/json"
	"github.com/go-chi/jwtauth/v5"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/errors"
	"github.com/yurchenkosv/gofermart/internal/model"
	"github.com/yurchenkosv/gofermart/internal/service"
	"io"
	"net/http"
)

type AuthHandler struct {
	authService service.Auth
	jwtAuth     *jwtauth.JWTAuth
}

func NewAuthHanler(authService *service.Auth, jwtAuth *jwtauth.JWTAuth) AuthHandler {
	return AuthHandler{
		authService: *authService,
		jwtAuth:     jwtAuth,
	}
}

func (h AuthHandler) HandleUserRegistration(writer http.ResponseWriter, request *http.Request) {
	user, err := h.parseForUser(request)
	if err != nil {
		log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	updatedUser, err := h.authService.RegisterUser(user)
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
	writer = SetToken(writer, *updatedUser, h.jwtAuth)
	writer.WriteHeader(http.StatusOK)
}

func (h AuthHandler) HanldeUserLogin(writer http.ResponseWriter, request *http.Request) {
	user, err := h.parseForUser(request)
	if err != nil {
		log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	updatedUser, err := h.authService.AuthenticateUser(user)
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
	writer = SetToken(writer, *updatedUser, h.jwtAuth)
	writer.WriteHeader(http.StatusOK)
}

func (h AuthHandler) parseForUser(request *http.Request) (*model.User, error) {
	var user model.User

	data, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &user)
	if err != nil {
		log.Error(err)
	}
	return &user, nil
}
