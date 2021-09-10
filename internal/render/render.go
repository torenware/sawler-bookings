package render

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/tsawler/bookings-app/internal/helpers"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"
	"github.com/tsawler/bookings-app/internal/config"
	"github.com/tsawler/bookings-app/internal/models"
)

// Create a function to render a time as a standard date
func standardDate(t time.Time) string {
	return t.Format("2006-01-02")
}

var functions = template.FuncMap{
	"stdDate": standardDate,
}

var app *config.AppConfig
var pathToTemplates = "./templates"

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// AddDefaultData adds data for all templates
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Authed = helpers.IsAuthed(r)
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.CSRFToken = nosurf.Token(r)
	return td
}

// fetchAndRenderTemplate factors out template rendering.
// r should be nil if not used to generate actual html pages.
func fetchAndRenderTemplate(tmpl string, td *models.TemplateData, r *http.Request) (*bytes.Buffer, error) {
	var tc map[string]*template.Template

	if app.UseCache {
		// get the template cache from the app config
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]
	if !ok {
		//log.Fatal("Could not get template from template cache")
		return nil, errors.New("could not get template from cache")
	}

	buf := new(bytes.Buffer)
    if r != nil {
		td = AddDefaultData(td, r)
	}

	err := t.Execute(buf, td)
	if err != nil {
		log.Printf("Error processing template: %s", err)
		return nil, err
	}
    return buf, nil
}

// TemplateAsString renders out a template as a string.
func TemplateAsString(tmpl string, td *models.TemplateData) (string, error) {
	buf, err := fetchAndRenderTemplate(tmpl, td, nil)
	if err != nil {
		log.Printf("Error processing template: %s", err)
		return "", err
	}
	if buf == nil {
		return "", errors.New("template buffer unavailable")
	}
    return buf.String(), nil
}

// Template renders a template
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
    buf, err := fetchAndRenderTemplate(tmpl, td, r)
	if err != nil {
		log.Printf("Error processing template: %s", err)
		return err
	}
	if buf == nil {
		return errors.New("template buffer unavailable")
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		fmt.Println("error writing template to browser", err)
		return err
	}
	return nil
}

// CreateTemplateCache creates a template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {

	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
