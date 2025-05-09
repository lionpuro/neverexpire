package main

import (
	"log"
	"net/http"

	"github.com/lionpuro/trackcert/views"
)

func (s *Server) handleHomePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		user, _ := s.Sessions.GetUser(r)
		if err := views.Home(w, user); err != nil {
			log.Printf("render template: %v", err)
		}
	}
}

func (s *Server) handleAccountPage(w http.ResponseWriter, r *http.Request) {
	user, err := s.Sessions.GetUser(r)
	if err != nil {
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
