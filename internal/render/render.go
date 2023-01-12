package render

import (
	"bytes"
	"fmt"
	"github.com/454270186/Hotel-booking-web-application/internal/Models"
	"github.com/454270186/Hotel-booking-web-application/internal/config"
	"github.com/justinas/nosurf"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var app *config.AppConfig

func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *Models.TemplateData, r *http.Request) *Models.TemplateData {
	td.CSRFToken = nosurf.Token(r) // make sure to go through the CSRF protection.
	// any post without this CSRFToken will be refused
	return td
}

// RenderTemplate is the Templates render
func RenderTemplate(w http.ResponseWriter, r *http.Request, html string, td *Models.TemplateData) {
	// create template cache from cache or create new
	var tc map[string]*template.Template
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	// get requested template from cache
	t, ok := tc[html]
	if !ok {
		log.Fatal("could not get template from templates cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	_ = t.Execute(buf, td)

	// render the template
	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser.", err)
	}
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// get all the file name "*.page.html" from ./templates
	pages, err := filepath.Glob("./templates/*.page.html")
	if err != nil {
		return myCache, err
	}

	// range through all files ending with *.page.html
	for _, page := range pages {
		name := filepath.Base(page) // get name like "*.page.html"
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// for layout
		matches, err := filepath.Glob("./templates/*.layout.html")
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.html")
			if err != nil {

				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil

}

/*
	RenderTemplate() is not very efficient. Everytime users load this page, template.ParseFiles()
parse files from the disk, with increment of the number of files, efficiency will become lower
and lower.

	To improve efficiency, Use a data structure to stored the template.
	Here use map to make a template cache(tc).
*/