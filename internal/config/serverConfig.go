package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/jwtauth/v5"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/dao"
)

type ServerConfig struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"0.0.0.0:8080"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	TokenAuth            *jwtauth.JWTAuth
	Repo                 *dao.PostgresRepository
}

func (config *ServerConfig) Parse() error {

	flag.StringVar(&config.RunAddress, "a", "0.0.0.0:8080", "-a <address>:<port>, default 0.0.0.0:8080")
	flag.StringVar(&config.DatabaseURI, "d", "", "-d <database uri>, postgresql://user:password@address:port/dbname")
	flag.StringVar(&config.AccrualSystemAddress, "r", "", "-r https://<address>:<port>")
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		log.Errorf("error when parse environment: %s", err)
		return err
	}
	return nil
}
