package handlers

import (
	"context"
	"github.com/go-chi/jwtauth/v5"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/config"
	"github.com/yurchenkosv/gofermart/internal/model"
	"net/http"
	"time"
)

func CheckErrors(err error, writer http.ResponseWriter) {
	if err != nil {
		log.Error("error processing request: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func GetConfigFromContext(ctx context.Context) config.ServerConfig {
	cfg := ctx.Value(model.ConfigKey("config")).(*config.ServerConfig)
	return *cfg
}

func SetToken(writer http.ResponseWriter, request *http.Request, user model.User) *http.ResponseWriter {
	claims := map[string]interface{}{
		"user_id": user.Id,
	}
	cfg := GetConfigFromContext(request.Context())
	tokenAuth := cfg.TokenAuth
	currentTime := time.Now()

	jwtauth.SetIssuedAt(claims, currentTime)
	jwtauth.SetExpiry(claims, currentTime.Add(5*time.Minute))
	_, token, _ := tokenAuth.Encode(claims)

	writer.Header().Add("jwt", token)

	return &writer
}

func GetUserIDFromToken(ctx context.Context) int {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := claims["user_id"].(float64)
	return int(userID)
}
