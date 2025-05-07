package main

import (
	"log"
	"net/http"

	"github.com/lionpuro/trackcert/views"
)

func handleHomePage(s *SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		user, _ := s.GetUser(r)
		if err := views.Home(w, user); err != nil {
			log.Printf(err.Error())
		}
	}
}

func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if err := views.Login(w); err != nil {
		log.Printf(err.Error())
	}
}
