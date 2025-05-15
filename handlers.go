package main

import (
	"log"
	"net/http"

	"github.com/lionpuro/trackcert/views"
)

func (s *Server) handleHomePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			handleErrorPage(w, r, "Page not found", http.StatusNotFound)
			return
		}
		u, ok := getUserCtx(r.Context())
		user := &u
		if !ok {
			user = nil
		}
		if err := views.Home(w, user); err != nil {
			log.Printf("render template: %v", err)
		}
	}
}

func (s *Server) handleAccountPage(w http.ResponseWriter, r *http.Request) {
	u, ok := getUserCtx(r.Context())
	user := &u
	if !ok {
		user = nil
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := views.Account(w, user); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if err := views.Login(w); err != nil {
		log.Printf(err.Error())
	}
}

func handleErrorPage(w http.ResponseWriter, r *http.Request, msg string, code int) {
	if err := views.Error(w, code, msg); err != nil {
		log.Printf("render template: %v", err)
	}
}
