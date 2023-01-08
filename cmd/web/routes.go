package main

import (
	"github.com/454270186/Hotel-booking-web-application/pkg/config"
	"github.com/454270186/Hotel-booking-web-application/pkg/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func routes(app *config.AppConfig) http.Handler {
	// go-pat
	//mux := pat.New()
	//
	//// set up the route
	//mux.Get("/", http.HandlerFunc(handler.Repo.Home))
	//mux.Get("/about", http.HandlerFunc(handler.Repo.About))

	// go-chi
	mux := chi.NewRouter()

	// use middleware
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// set up routes
	mux.Get("/", handler.Repo.Home)
	mux.Get("/about", handler.Repo.About)

	return mux
}
