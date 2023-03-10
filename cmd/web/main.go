package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/454270186/Hotel-booking-web-application/internal/Models"
	"github.com/454270186/Hotel-booking-web-application/internal/config"
	"github.com/454270186/Hotel-booking-web-application/internal/driver"
	"github.com/454270186/Hotel-booking-web-application/internal/handler"
	"github.com/454270186/Hotel-booking-web-application/internal/helpers"
	"github.com/454270186/Hotel-booking-web-application/internal/render"
	"github.com/alexedwards/scs/v2"
	"log"
	"net/http"
	"os"
	"time"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)
	fmt.Println("Starting mail listener...")
	listenForMail()

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

func run() (*driver.DB, error) {
	// What I am to store in Session
	gob.Register(Models.Reservation{})
	gob.Register(Models.User{})
	gob.Register(Models.Room{})
	gob.Register(Models.Restriction{})
	gob.Register(Models.RoomRestriction{})
	gob.Register(map[string]int{})

	// read flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPass := flag.String("dbpass", "", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()
	if *dbName == "" || *dbUser == "" || *dbPass == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}

	mailChan := make(chan Models.MailData)
	app.MailChan = mailChan

	// change this to true when in production
	app.InProduction = *inProduction
	app.UseCache = *useCache

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	// connect to database
	log.Println("Connect to database...")
	connectionStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)
	db, err := driver.ConnectSQL(connectionStr)
	if err != nil {
		log.Fatal("Cannot connect to database, Dying...")
	}
	log.Println("Connected to database")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	app.TemplateCache = tc

	repo := handler.NewRepo(&app, db)
	handler.NewHandler(repo) // new and set repository for handler
	render.NewRenderer(&app) // new and set up app config for render
	helpers.NewHelpers(&app) // new and set up app config for helpers

	return db, nil
}
