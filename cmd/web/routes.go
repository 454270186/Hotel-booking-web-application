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

	mux.Get("/choose-room/{id}", handler.Repo.ChooseRoom)
	mux.Get("/book-room", handler.Repo.BookRoom)

	mux.Get("/contact", handler.Repo.Contact)

	mux.Get("/make-reservation", handler.Repo.Reservation)
	mux.Post("/make-reservation", handler.Repo.PostReservation)
	mux.Get("/reservation-summary", handler.Repo.ReservationSummary)

	mux.Get("/user/login", handler.Repo.ShowLogin)
	mux.Post("/user/login", handler.Repo.PostShowLogin)
	mux.Get("/user/logout", handler.Repo.Logout)

	// 处理静态文件，让网页可以访问到static文件夹里的文件
	// 这一步非常重要！！
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(mux chi.Router) {
		//mux.Use(Auth)

		mux.Get("/dashboard", handler.Repo.AdminDashboard)

		mux.Get("/reservations-new", handler.Repo.AdminNewReservations)
		mux.Get("/reservations-all", handler.Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", handler.Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", handler.Repo.AdminPostReservationsCalendar)

		mux.Get("/process-reservation/{src}/{id}/do", handler.Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}/do", handler.Repo.AdminDeleteReservation)

		mux.Get("/reservations/{src}/{id}/show", handler.Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handler.Repo.AdminPostShowReservation)
	})

	return mux
}
