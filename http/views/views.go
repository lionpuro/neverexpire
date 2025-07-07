package views

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"path/filepath"

	"github.com/lionpuro/neverexpire/model"
	"github.com/lionpuro/neverexpire/notification"
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
	homeTmpl       = parse("pages/index.html")
	errorPageTmpl  = parse("pages/error.html")
	domainsTmpl    = parse("pages/domains/domains.html")
	domainTmpl     = parse("pages/domains/domain.html")
	newDomainsTmpl = parse("pages/domains/new.html")
	settingsTmpl   = parse("pages/settings.html")
	loginTmpl      = parse("pages/login.html")
	partials       = parsePartials()
)

func Home(w io.Writer, ld LayoutData) error {
	return homeTmpl.render(w, map[string]any{"LayoutData": ld})
}

func Error(w io.Writer, ld LayoutData, code int, msg string) error {
	data := map[string]any{"LayoutData": ld, "Code": code, "Message": msg}
	return errorPageTmpl.render(w, data)
}

func Domains(w io.Writer, ld LayoutData, domains []model.Domain) error {
	data := map[string]any{"LayoutData": ld, "Domains": domains}
	return domainsTmpl.render(w, data)
}

func Domain(w io.Writer, ld LayoutData, d model.Domain) error {
	return domainTmpl.render(w, map[string]any{"LayoutData": ld, "Domain": d})
}

func NewDomains(w io.Writer, ld LayoutData, inputValue string) error {
	data := map[string]any{"LayoutData": ld, "InputValue": inputValue}
	if inputValue == "" {
		data["InputValue"] = nil
	}
	return newDomainsTmpl.render(w, data)
}

func Settings(w io.Writer, ld LayoutData, sett model.Settings) error {
	type reminder struct {
		Value   int
		Display string
	}
	opts := []reminder{
		{Value: notification.ThresholdDay, Display: "1 day before"},
		{Value: notification.Threshold2Days, Display: "2 days before"},
		{Value: notification.ThresholdWeek, Display: "1 week before"},
		{Value: notification.Threshold2Weeks, Display: "2 weeks before"},
	}
	data := map[string]any{
		"LayoutData":      ld,
		"ReminderOptions": opts,
		"Settings":        sett,
	}
	return settingsTmpl.render(w, data)
}

func Login(w io.Writer) error {
	return loginTmpl.render(w, nil)
}

func ErrorBanner(w io.Writer, err error) error {
	return partials.renderPartial(w, "error-banner", map[string]any{"Error": err})
}

func SuccessBanner(w io.Writer, msg string) error {
	return partials.renderPartial(w, "success-banner", map[string]any{"Message": msg})
}

func parse(templates ...string) *viewTemplate {
	funcs := funcMap()
	patterns := []string{templatePath("components/*.html")}
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

func (t *viewTemplate) renderPartial(w io.Writer, name string, data any) error {
	buf := &bytes.Buffer{}
	if err := t.template.ExecuteTemplate(buf, name, data); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}

func (t *viewTemplate) render(w io.Writer, data any) error {
	buf := &bytes.Buffer{}
	if err := t.template.Execute(buf, data); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}

func templatePath(file string) string {
	return filepath.Join("templates", file)
}
