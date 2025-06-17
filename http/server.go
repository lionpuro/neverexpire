package http

import (
	"net/http"
)

func NewServer(h *Handler) *http.Server {
	r := http.NewServeMux()

	handle := func(p string, hf http.HandlerFunc) {
		r.HandleFunc(p, h.Authenticate(hf))
	}

	handle("GET /", h.HomePage)
	handle("GET /domains", h.RequireAuth(h.DomainsPage))
	handle("GET /domains/new", h.RequireAuth(h.NewDomainPage))
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

	r.HandleFunc("GET /auth/google/login", h.Login(h.AuthService.GoogleClient))
	r.HandleFunc("GET /auth/google/callback", h.AuthCallback(h.AuthService.GoogleClient))
	r.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("assets/public"))))

	srv := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	return srv
}
