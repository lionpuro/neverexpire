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
	home          = parse("layouts/main.html", "home.html")
	errorPage     = parse("layouts/main.html", "error.html")
	domains       = parse("layouts/main.html", "domains/index.html")
	domain        = parse("layouts/main.html", "domains/details.html")
	newDomain     = parse("layouts/main.html", "domains/new.html")
	settings      = parse("layouts/main.html", "settings.html")
	login         = parse("layouts/auth.html", "login.html")
	errorBanner   = parse("partials/error-banner.html")
	successBanner = parse("partials/success-banner.html")
	domainPart    = parse("layouts/partial.html", "domains/details.html")
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

func Domain(w http.ResponseWriter, user *model.User, d model.Domain, err error, refreshData bool) error {
	return domain.render(w, map[string]any{"User": user, "Domain": d, "Error": err, "RefreshData": refreshData})
}

func DomainPartial(w http.ResponseWriter, d model.Domain) error {
	return domainPart.render(w, map[string]any{"Domain": d})
}

func NewDomain(w http.ResponseWriter, user *model.User, inputValue string, err error) error {
	data := map[string]any{"User": user, "InputValue": inputValue, "Error": err}
	if inputValue == "" {
		data["InputValue"] = nil
	}
	return newDomain.render(w, data)
}

func Settings(w http.ResponseWriter, user *model.User, sett model.Settings) error {
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
		"User":            user,
		"ReminderOptions": opts,
		"Settings":        sett,
	}
	return settings.render(w, data)
}

func Login(w http.ResponseWriter) error {
	return login.render(w, nil)
}

func ErrorBanner(w http.ResponseWriter, err error) error {
	return errorBanner.render(w, map[string]any{"Error": err})
}

func SuccessBanner(w http.ResponseWriter, msg string) error {
	return successBanner.render(w, map[string]any{"Message": msg})
}

func parse(templates ...string) *viewTemplate {
	funcs := template.FuncMap{
		"datef":          datef,
		"sprintf":        fmt.Sprintf,
		"cn":             cn,
		"statusClass":    statusClass,
		"statusText":     statusText,
		"withAttributes": withAttributes,
		"split":          split,
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
