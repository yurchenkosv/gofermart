package main

import (
	"github.com/golang-migrate/migrate/v4"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/config"
)

func Migrate(cfg *config.ServerConfig) {
	m, err := migrate.New(
		"file://db/migrations",
		cfg.DatabaseURI)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

}
