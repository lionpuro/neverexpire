package views

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/lionpuro/neverexpire/model"
)

//go:embed templates
var templateFS embed.FS

type viewTemplate struct {
	template *template.Template
}

type LayoutData struct {
	User  *model.User
	Error error
}

var (
	home      = parse("layouts/main.html", "home.html")
	errorPage = parse("layouts/main.html", "error.html")
	domains   = parse("layouts/main.html", "domains/index.html")
	domain    = parse("layouts/main.html", "domains/details.html")
	newDomain = parse("layouts/main.html", "domains/new.html")
	settings  = parse("layouts/main.html", "settings.html")
	login     = parse("layouts/main.html", "login.html")
	partials  = parsePartials()
)

func Home(w http.ResponseWriter, ld LayoutData) error {
	return home.render(w, map[string]any{"LayoutData": ld})
}

func Error(w http.ResponseWriter, ld LayoutData, code int, msg string) error {
	data := map[string]any{"LayoutData": ld, "Code": code, "Message": msg}
	return errorPage.render(w, data)
}

func Domains(w http.ResponseWriter, ld LayoutData, dmains []model.Domain) error {
	data := map[string]any{"LayoutData": ld, "Domains": dmains}
	return domains.render(w, data)
}

func Domain(w http.ResponseWriter, ld LayoutData, d model.Domain) error {
	return domain.render(w, map[string]any{"LayoutData": ld, "Domain": d})
}

func NewDomain(w http.ResponseWriter, ld LayoutData, inputValue string) error {
	data := map[string]any{"LayoutData": ld, "InputValue": inputValue}
	if inputValue == "" {
		data["InputValue"] = nil
	}
	return newDomain.render(w, data)
}

func Settings(w http.ResponseWriter, ld LayoutData, sett model.Settings) error {
	type reminder struct {
		Value   int
		Display string
	}
	day := 24 * 60 * 60
	opts := []reminder{
		{Value: 1 * day, Display: "1 day before"},
		{Value: 2 * day, Display: "2 days before"},
		{Value: 7 * day, Display: "1 week before"},
		{Value: 14 * day, Display: "2 weeks before"},
	}
	data := map[string]any{
		"LayoutData":      ld,
		"ReminderOptions": opts,
		"Settings":        sett,
	}
	return settings.render(w, data)
}

func Login(w http.ResponseWriter) error {
	return login.render(w, nil)
}

func ErrorBanner(w http.ResponseWriter, err error) error {
	return partials.renderPartial(w, "error-banner", map[string]any{"Error": err})
}

func SuccessBanner(w http.ResponseWriter, msg string) error {
	return partials.renderPartial(w, "success-banner", map[string]any{"Message": msg})
}

func parse(templates ...string) *viewTemplate {
	funcs := funcMap()
	patterns := []string{templatePath("base.html"), templatePath("components/*.html")}
	for _, t := range templates {
		patterns = append(patterns, templatePath(t))
	}
	name := filepath.Base(templates[0])
	tmpl := template.Must(template.New(name).Funcs(funcs).ParseFS(templateFS, patterns...))
	return &viewTemplate{template: tmpl}
}

func parsePartials() *viewTemplate {
	funcs := funcMap()
	patterns := []string{templatePath("components/*.html")}
	tmpl := template.Must(template.New("").Funcs(funcs).ParseFS(templateFS, patterns...))
	return &viewTemplate{template: tmpl}
}

func (t *viewTemplate) renderPartial(w http.ResponseWriter, name string, data any) error {
	buf := &bytes.Buffer{}
	if err := t.template.ExecuteTemplate(buf, name, data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := buf.WriteTo(w)
	return err
}

func (t *viewTemplate) render(w http.ResponseWriter, data any) error {
	buf := &bytes.Buffer{}
	if err := t.template.Execute(buf, data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := buf.WriteTo(w)
	return err
}

func templatePath(file string) string {
	return filepath.Join("templates", file)
}
