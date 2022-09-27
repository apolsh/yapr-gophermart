package http

import (
	"net/http"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/service"
	"github.com/go-chi/chi/v5"
)

type ordersRoutes struct {
	gophermartService service.GophermartService
}

func newOrdersRoutes(router *chi.Mux, s service.GophermartService) {
	o := &ordersRoutes{gophermartService: s}

	router.Route("/api/user", func(router chi.Router) {
		router.Group(func(r chi.Router) {
			//set auth middleware
			router.Use(AuthMiddleware(s.ParseJWTToken))
			router.Route("/order", func(router chi.Router) {
				router.Post("/", o.createOrder)
				router.Get("/", o.getOrders)
			})
			router.Route("/balance", func(router chi.Router) {
				router.Get("/", o.getBalance)
				router.Post("/withdraw", o.createWithdraw)
			})
			router.Get("/withdrawals", o.getWithdrawals)
		})

	})
}

func (o *ordersRoutes) createOrder(w http.ResponseWriter, r *http.Request) {
	//TODO: implement me
}

func (o *ordersRoutes) getOrders(w http.ResponseWriter, r *http.Request) {
	//TODO: implement me
}

func (o *ordersRoutes) getBalance(w http.ResponseWriter, r *http.Request) {
	//TODO: implement me
}

func (o *ordersRoutes) createWithdraw(w http.ResponseWriter, r *http.Request) {
	//TODO: implement me
}

func (o *ordersRoutes) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	//TODO: implement me
}
