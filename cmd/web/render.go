package main

import (
	"fmt"
	"net/http"
	"text/template"
	"time"
)

var pathToTemplate = "./cmd/web/templates"

type TemplateDate struct {
	StringMap     map[string]string
	IntMAp        map[string]int
	FloatMap      map[string]float64
	Data          map[string]any
	Flash         string
	Warning       string
	Error         string
	Authenticated bool
	Now           time.Time
	//User *data.User
}

func (app *Config) render(w http.ResponseWriter, r *http.Request, t string, td *TemplateDate) {
	partials := []string{
		fmt.Sprintf("%s/base.layout.gohtml", pathToTemplate),
		fmt.Sprintf("%s/header.partial.gohtml", pathToTemplate),
		fmt.Sprintf("%s/navbar.partial.gohtml", pathToTemplate),
		fmt.Sprintf("%s/footer.partial.gohtml", pathToTemplate),
		fmt.Sprintf("%s/alerts.partial.gohtml", pathToTemplate),
	}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf("%s/%s", pathToTemplate, t))

	for _, x := range partials {
		templateSlice = append(templateSlice, x)
	}

	if td == nil {
		td = &TemplateDate{}
	}

	temp, err := template.ParseFiles(templateSlice...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := temp.Execute(w, app.AddDefaultData(td, r)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *Config) AddDefaultData(td *TemplateDate, r *http.Request) *TemplateDate {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	if app.IsAuthenticated(r) {
		td.Authenticated = true
	}
	td.Now = time.Now()
	return td
}

func (app *Config) IsAuthenticated(r *http.Request) bool {
	return app.Session.Exists(r.Context(), "userID")
}
