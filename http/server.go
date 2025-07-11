package http

import (
	"fmt"
	"net/http"

	"github.com/lionpuro/neverexpire/api"
)

func NewServer(port int, h *Handler, ah *api.Handler) *http.Server {
	r := http.NewServeMux()

	web := func(p string, hf http.HandlerFunc) {
		r.Handle(p, contentType(h.Authenticate(hf), "text/html; charset=utf-8"))
	}
	api := func(p string, hf http.HandlerFunc) {
		r.Handle(p, contentType(ah.RequireAuth(hf), "application/json"))
	}

	// web
	web("GET /", h.HomePage)
	web("GET /domains", h.RequireAuth(h.DomainsPage))
	web("GET /domains/new", h.RequireAuth(h.NewDomainsPage))
	web("POST /domains", h.RequireAuth(h.CreateDomains))
	web("GET /domains/{id}", h.RequireAuth(h.DomainPage))
	web("DELETE /domains/{id}", h.RequireAuth(h.DeleteDomain))
	web("GET /login", h.LoginPage)
	web("GET /logout", h.Logout)
	web("DELETE /account", h.RequireAuth(h.DeleteAccount))
	web("GET /settings", h.RequireAuth(h.SettingsPage))
	web("PUT /settings/reminders", h.RequireAuth(h.UpdateReminders))
	web("POST /settings/webhook", h.RequireAuth(h.AddWebhook))
	web("DELETE /settings/webhook", h.RequireAuth(h.DeleteWebhook))
	web("GET /api", h.RequireAuth(h.APIPage))
	web("GET /account/tokens/new", h.RequireAuth(h.CreateAPIKey))
	web("DELETE /account/tokens/{id}", h.RequireAuth(h.DeleteAPIKey))
	r.HandleFunc("GET /auth/google/login", h.Login(h.AuthService.GoogleClient))
	r.HandleFunc("GET /auth/google/callback", h.AuthCallback(h.AuthService.GoogleClient))
	r.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("assets/public"))))
	// api
	api("GET /api/hosts", ah.Hosts)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
	return srv
}
