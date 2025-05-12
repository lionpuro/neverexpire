package views

import (
	"bytes"
	"embed"
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
	home      = parseTemplate("home.html", "layouts/main.html")
	errorPage = parseTemplate("error.html", "layouts/main.html")
	account   = parseTemplate("account.html", "layouts/main.html")
	login     = parseTemplate("login.html", "layouts/auth.html")
)

func Home(w http.ResponseWriter, user *model.SessionUser) error {
	return home.render(w, map[string]any{"User": user})
}

func Error(w http.ResponseWriter, code int, msg string) error {
	return errorPage.render(w, map[string]any{"User": nil, "Code": code, "Message": msg})
}

func Account(w http.ResponseWriter, user *model.SessionUser) error {
	return account.render(w, map[string]any{"User": user})
}

func Login(w http.ResponseWriter) error {
	return login.render(w, nil)
}

func parseTemplate(templates ...string) *viewTemplate {
	patterns := []string{templatePath("base.html")}
	for _, t := range templates {
		patterns = append(patterns, templatePath(t))
	}
	tmpl := template.Must(template.New("base.html").ParseFS(templateFS, patterns...))
	tmpl = template.Must(tmpl.ParseFS(templateFS, templatePath("components/*.html")))
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
