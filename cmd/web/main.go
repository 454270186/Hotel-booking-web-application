package main

import (
	"fmt"
	"github.com/454270186/Hotel-booking-web-application/pkg/config"
	"github.com/454270186/Hotel-booking-web-application/pkg/handler"
	"github.com/454270186/Hotel-booking-web-application/pkg/render"
	"github.com/alexedwards/scs/v2"
	"log"
	"net/http"
	"time"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

func main() {
	// change this to true when in production
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	app.TemplateCache = tc
	app.UseCache = false // do not use the Template cache, render from disk
	render.NewTemplates(&app)

	// New and set repository for handler
	var repo *handler.Repository
	repo = handler.NewRepo(&app)
	handler.NewHandler(repo)

	//http.HandleFunc("/", handler.Repo.Home)
	//http.HandleFunc("/about", handler.Repo.About)

	fmt.Printf("Starting application on port: %s\n", portNumber)

	//_ = http.ListenAndServe(portNumber, nil)

	server := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = server.ListenAndServe()
	log.Fatal(err)

}
