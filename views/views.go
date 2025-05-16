package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/lionpuro/trackcerts/model"
)

//go:embed templates
var templateFS embed.FS

type viewTemplate struct {
	template *template.Template
}

var (
	home        = parse("layouts/main.html", "home.html")
	errorPage   = parse("layouts/main.html", "error.html")
	domains     = parse("layouts/main.html", "domains/index.html")
	domain      = parse("layouts/main.html", "domains/details.html")
	newDomain   = parse("layouts/main.html", "domains/new.html")
	account     = parse("layouts/main.html", "account.html")
	login       = parse("layouts/auth.html", "login.html")
	errorBanner = parse("partials/error-banner.html")
)

func Home(w http.ResponseWriter, user *model.User, err error) error {
	return home.render(w, map[string]any{"Error": err, "User": user})
}

func Error(w http.ResponseWriter, code int, msg string) error {
	return errorPage.render(w, map[string]any{"User": nil, "Code": code, "Message": msg})
}

func Domains(w http.ResponseWriter, user *model.User, dmains []model.Domain, err error) error {
	return domains.render(w, map[string]any{"User": user, "Domains": dmains, "Error": err})
}

func Domain(w http.ResponseWriter, user *model.User, d model.Domain, err error) error {
	return domain.render(w, map[string]any{"User": user, "Domain": d, "Error": err})
}

func NewDomain(w http.ResponseWriter, user *model.User, inputValue string, err error) error {
	data := map[string]any{"User": user, "InputValue": inputValue, "Error": err}
	if inputValue == "" {
		data["InputValue"] = nil
	}
	return newDomain.render(w, data)
}

func Account(w http.ResponseWriter, user *model.User) error {
	return account.render(w, map[string]any{"User": user})
}

func Login(w http.ResponseWriter) error {
	return login.render(w, nil)
}

func ErrorBanner(w http.ResponseWriter, err error) error {
	return errorBanner.render(w, map[string]any{"Error": err})
}

func parse(templates ...string) *viewTemplate {
	funcs := template.FuncMap{
		"datef":          datef,
		"sprintf":        fmt.Sprintf,
		"cn":             cn,
		"statusClass":    statusClass,
		"statusText":     statusText,
		"withAttributes": withAttributes,
	}
	patterns := []string{templatePath("base.html"), templatePath("components/*.html")}
	for _, t := range templates {
		patterns = append(patterns, templatePath(t))
	}
	name := filepath.Base(templates[0])
	tmpl := template.Must(template.New(name).Funcs(funcs).ParseFS(templateFS, patterns...))
	return &viewTemplate{template: tmpl}
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
