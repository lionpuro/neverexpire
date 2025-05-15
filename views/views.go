package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/lionpuro/trackcert/model"
)

//go:embed templates
var templateFS embed.FS

type viewTemplate struct {
	template *template.Template
}

var (
	home      = parse("layouts/main.html", "home.html")
	errorPage = parse("layouts/main.html", "error.html")
	domains   = parse("layouts/main.html", "domains/index.html")
	domain    = parse("layouts/main.html", "domains/details.html")
	newDomain = parse("layouts/main.html", "domains/new.html")
	account   = parse("layouts/main.html", "account.html")
	login     = parse("layouts/auth.html", "login.html")
)

func Home(w http.ResponseWriter, user *model.User) error {
	return home.render(w, map[string]any{"User": user})
}

func Error(w http.ResponseWriter, code int, msg string) error {
	return errorPage.render(w, map[string]any{"User": nil, "Code": code, "Message": msg})
}

func Domains(w http.ResponseWriter, user *model.User, dmains []model.Domain) error {
	return domains.render(w, map[string]any{"User": user, "Domains": dmains})
}

func Domain(w http.ResponseWriter, user *model.User, d model.Domain) error {
	return domain.render(w, map[string]any{"User": user, "Domain": d})
}

func NewDomain(w http.ResponseWriter, user *model.User) error {
	return newDomain.render(w, map[string]any{"User": user})
}

func Account(w http.ResponseWriter, user *model.User) error {
	return account.render(w, map[string]any{"User": user})
}

func Login(w http.ResponseWriter) error {
	return login.render(w, nil)
}

func parse(templates ...string) *viewTemplate {
	funcs := template.FuncMap{
		"datef":          datef,
		"sprintf":        fmt.Sprintf,
		"cn":             cn,
		"statusClass":    statusClass,
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
