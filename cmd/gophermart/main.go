package main

import (
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-co-op/gocron"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	log "github.com/sirupsen/logrus"
	"github.com/yurchenkosv/gofermart/internal/config"
	"github.com/yurchenkosv/gofermart/internal/controllers"
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/routers"
	"net/http"
	"os"
	"time"
)

var (
	cfg       = &config.ServerConfig{}
	tokenAuth *jwtauth.JWTAuth
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func main() {
	token, err := generateRandomToken()
	if err != nil {
		log.Fatal(err)
	}
	err = cfg.Parse()
	if err != nil {
		log.Error(err)
	}
	tokenAuth = jwtauth.New("HS256", token, nil)
	cfg.TokenAuth = tokenAuth
	cfg.Repo = dao.NewPGRepo(cfg.DatabaseURI)
	osSignal := make(chan os.Signal, 1)

	Migrate(cfg)
	router := routers.NewRouter(cfg)
	server := http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	sched := gocron.NewScheduler(time.UTC)
	sched.EveryRandom(2, 7).
		Second().
		Do(controllers.StatusCheckLoop, cfg)
	sched.StartAsync()

	go func() {
		<-osSignal
		sched.Stop()
		os.Exit(0)
	}()
	log.Fatal(server.ListenAndServe())

}
