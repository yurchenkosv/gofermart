package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/yurchenkosv/gofermart/internal/config"
	"github.com/yurchenkosv/gofermart/internal/handlers"
	"github.com/yurchenkosv/gofermart/internal/middlewares"
)

func NewRouter(cfg *config.ServerConfig) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.RequestID)
	router.Use(middleware.StripSlashes)
	router.Use(middlewares.AppendConfigToContext(cfg))

	router.Route("/api/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlewares.AllowContentType("application/json"))
			r.Post("/register", handlers.HandleUserRegistration)
			r.Post("/login", handlers.HandleUserLogin)
		})
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(cfg.TokenAuth))
			r.Use(jwtauth.Authenticator)
			r.Group(func(r chi.Router) {
				r.Use(middlewares.AllowContentType("text/plain"))
				r.Post("/orders", handlers.HandleCreateOrder)
			})
			r.Get("/orders", handlers.HandleGetOrders)
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", handlers.HandleGetBalance)
				r.Post("/withdraw", handlers.HandleBalanceDraw)
				r.Get("/withdrawals", handlers.HandleGetBalanceDraws)
			})
		})
	})

	return router
}
