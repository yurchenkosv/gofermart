package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/jwtauth/v5"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/dao"
)

type ServerConfig struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	TokenAuth            *jwtauth.JWTAuth
	Repo                 *dao.PostgresRepository
}

func (config *ServerConfig) Parse() error {
	err := env.Parse(config)
	if err != nil {
		log.Errorf("error when parse environment: %s", err)
		return err
	}
	return nil
}
