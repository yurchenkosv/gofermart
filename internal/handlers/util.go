package handlers

import (
	"context"
	"github.com/go-chi/jwtauth/v5"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/model"
	"net/http"
	"time"
)

func SetToken(writer http.ResponseWriter, user model.User, auth *jwtauth.JWTAuth) http.ResponseWriter {
	claims := map[string]interface{}{
		"user_id": *user.ID,
	}
	currentTime := time.Now()

	jwtauth.SetIssuedAt(claims, currentTime)
	jwtauth.SetExpiry(claims, currentTime.Add(5*time.Minute))
	_, token, err := auth.Encode(claims)

	if err != nil {
		log.Error("error setting token for user:", err)
		return writer
	}

	writer.Header().Add("jwt", token)
	writer.Header().Add("Set-Cookie", "jwt="+token)
	return writer
}

func GetUserIDFromToken(ctx context.Context) int {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		log.Error(err)
	}
	userID := claims["user_id"].(float64)
	return int(userID)
}
