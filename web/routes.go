package web

import (
	"net/http"

	"github.com/lionpuro/neverexpire/web/handlers"
)

func NewRouter(h *handlers.Handler) *http.ServeMux {
	r := http.NewServeMux()

	handle := func(p string, hf http.HandlerFunc) {
		r.Handle(p, h.Authenticate(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				hf(w, r)
			}),
		))
	}

	handle("GET /{$}", h.HomePage)
	handle("GET /domains", h.RequireAuth(h.DomainsPage))
	handle("GET /domains/new", h.RequireAuth(h.NewDomainsPage))
	handle("POST /domains", h.RequireAuth(h.CreateDomains))
	handle("GET /domains/{id}", h.RequireAuth(h.DomainPage))
	handle("DELETE /domains/{id}", h.RequireAuth(h.DeleteDomain))
	handle("GET /login", h.LoginPage)
	handle("GET /logout", h.Logout)
	handle("DELETE /account", h.RequireAuth(h.DeleteAccount))
	handle("GET /settings", h.RequireAuth(h.SettingsPage))
	handle("PUT /settings/reminders", h.RequireAuth(h.UpdateReminders))
	handle("POST /settings/webhook", h.RequireAuth(h.AddWebhook))
	handle("DELETE /settings/webhook", h.RequireAuth(h.DeleteWebhook))
	handle("GET /account/api", h.RequireAuth(h.APIPage))
	handle("GET /account/tokens/new", h.RequireAuth(h.CreateAPIKey))
	handle("DELETE /account/tokens/{id}", h.RequireAuth(h.DeleteAPIKey))
	r.HandleFunc("GET /auth/google/login", h.Login(h.AuthService.GoogleClient))
	r.HandleFunc("GET /auth/google/callback", h.AuthCallback(h.AuthService.GoogleClient))
	r.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("assets/public"))))

	return r
}
