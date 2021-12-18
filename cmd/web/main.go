package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/yalagtyarzh/leafsite/internal/config"
	"github.com/yalagtyarzh/leafsite/internal/handlers"
	"github.com/yalagtyarzh/leafsite/internal/models"
	"github.com/yalagtyarzh/leafsite/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

//main is the main application function
func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting application on port %s", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() error {
	//What am I going to put in the session
	gob.Register(models.Reservation{})

	//Change this to true when in production
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		return errors.New("cannot create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.NewTemplates(&app)

	return nil
}
