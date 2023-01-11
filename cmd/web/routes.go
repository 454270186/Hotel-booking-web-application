package main

import (
	"github.com/454270186/Hotel-booking-web-application/internal/config"
	"github.com/454270186/Hotel-booking-web-application/internal/handler"
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
	mux.Get("/generals-quarters", handler.Repo.Generals)
	mux.Get("/majors-suite", handler.Repo.Majors)

	mux.Get("/search-availability", handler.Repo.Availability)
	mux.Post("/search-availability", handler.Repo.PostAvailability)
	mux.Post("/search-availability-json", handler.Repo.AvailabilityJSON)

	mux.Get("/contact", handler.Repo.Contact)

	mux.Get("/make-reservation", handler.Repo.Reservation)
	mux.Post("/make-reservation", handler.Repo.PostReservation)

	// 处理静态文件，让网页可以访问到static文件夹里的文件
	// 这一步非常重要！！
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}
