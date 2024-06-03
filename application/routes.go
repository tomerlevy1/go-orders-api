package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tomerlevy1/go-orders-api/handler"
	"github.com/tomerlevy1/go-orders-api/repository/order"
)

func (app *App) loadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/orders", app.loadOrderRoutes)

	app.router = router
}

func (app *App) loadOrderRoutes(router chi.Router) {
	orderHandler := &handler.Order{
		Repo: &order.RedisRepo{
			Client: app.rdb,
		},
	}

	router.Get("/", orderHandler.List)
	router.Get("/{id}", orderHandler.GetById)
	router.Post("/", orderHandler.Create)
	router.Put("/{id}", orderHandler.UpdateById)
	router.Delete("/{id}", orderHandler.DeleteById)
}
