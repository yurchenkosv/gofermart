package main

import (
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/config"
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/routers"
	"net/http"
)

var (
	cfg       = config.ServerConfig{}
	tokenAuth *jwtauth.JWTAuth
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	err := cfg.Parse()
	if err != nil {
		log.Error(err)
	}
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
	cfg.TokenAuth = tokenAuth
	cfg.Repo = dao.NewPGRepo(cfg.DatabaseURI)

	Migrate(&cfg)
	router := routers.NewRouter(&cfg)

	server := http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}
	log.Fatal(server.ListenAndServe())
}
