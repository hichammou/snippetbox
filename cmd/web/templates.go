package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"snippetbox.hichammou/internal/models"
	"snippetbox.hichammou/ui"
)

// Define a templateData type to act as the holding structure formaing
// any dynamic data that we want to pass to our HTML templates.
type templateData struct {
	CurrentYear     int
	Snippet         models.Snippet
	Snippets        []models.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")

	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		// extract the file name from the full path
		name := filepath.Base(page)

		// create a slice containing the filepath patterns for the template we want to parse
		patterns := []string{
			"html/base.html",
			"html/partials/*.html",
			page,
		}

		// Parse the base template into a template set.
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
