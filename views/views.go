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
	home  = parseTemplate("home.html")
	login = parseTemplate("login.html")
)

func Home(w http.ResponseWriter, user *model.SessionUser) error {
	return home.render(w, map[string]any{"User": user})
}

func Login(w http.ResponseWriter) error {
	return login.render(w, nil)
}

func parseTemplate(name string) *viewTemplate {
	tmpl := template.Must(template.New("base.html").ParseFS(templateFS, templatePath("base.html"), templatePath(name)))
	tmpl = template.Must(tmpl.ParseFS(templateFS, templatePath("components/*.html"), templatePath("layouts/*.html")))
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
