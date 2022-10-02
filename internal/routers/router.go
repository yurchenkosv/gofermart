package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/yurchenkosv/gofermart/internal/dao"
	"github.com/yurchenkosv/gofermart/internal/handlers"
	"github.com/yurchenkosv/gofermart/internal/middlewares"
	"github.com/yurchenkosv/gofermart/internal/service"
)

func NewRouter(repo dao.Repository, tokenAuth *jwtauth.JWTAuth) chi.Router {
	var (
		authService     = service.NewAuthService(repo)
		orderService    = service.NewOrderService(repo)
		withdrawService = service.NewWithdrawService(repo)
		balanceService  = service.NewBalance(repo)

		authHandler    = handlers.NewAuthHanler(&authService, tokenAuth)
		orderHandler   = handlers.NewOrderHandler(&orderService)
		balanceHandler = handlers.NewBalanceHandler(&balanceService, &withdrawService)
	)

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.RequestID)
	router.Use(middleware.StripSlashes)

	router.Route("/api/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlewares.AllowContentType("application/json"))
			r.Post("/register", authHandler.HandleUserRegistration)
			r.Post("/login", authHandler.HanldeUserLogin)
		})
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator)
			r.Group(func(r chi.Router) {
				r.Use(middlewares.AllowContentType("text/plain"))
				r.Post("/orders", orderHandler.HandleCreateOrder)
			})
			r.Use(middlewares.AllowContentType("application/json"))
			r.Get("/orders", orderHandler.HandleGetOrders)
			r.Get("/withdrawals", balanceHandler.HandleGetBalanceWithdraws)
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", balanceHandler.HandleGetBalance)
				r.Post("/withdraw", balanceHandler.HandleBalanceWithdraw)
			})
		})
	})

	return router
}
