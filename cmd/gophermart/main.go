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
	"os/signal"
	"syscall"
	"time"
)

var (
	cfg       = &config.ServerConfig{}
	tokenAuth *jwtauth.JWTAuth
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.WarnLevel)
}

func main() {
	err := cfg.Parse()
	if err != nil {
		log.Error(err)
	}
	tokenAuth = jwtauth.New("HS256", []byte(cfg.InitialTokenSecret), nil)
	cfg.TokenAuth = tokenAuth
	cfg.Repo = dao.NewPGRepo(cfg.DatabaseURI)
	cfg.Repo.Migrate("file://db/migrations")

	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	sched := gocron.NewScheduler(time.UTC)
	_, err = sched.EveryRandom(2, 7).
		Second().
		Do(controllers.StatusCheckLoop, cfg)
	if err != nil {
		log.Fatal("cannot create scheduler for update tasks", err)
	}
	sched.StartAsync()

	router := routers.NewRouter(cfg)
	server := http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	go func() {
		<-osSignal
		sched.Stop()
		cfg.Repo.Shutdown()
		os.Exit(0)
	}()

	log.Fatal(server.ListenAndServe())

}
