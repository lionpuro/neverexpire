package main

import (
	"github.com/lionpuro/trackcerts/model"
	"github.com/lionpuro/trackcerts/user"
	"github.com/lionpuro/trackcerts/views"
	"log"
	"net/http"
)

func handleHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		if err := views.Error(w, http.StatusNotFound, "Page not found"); err != nil {
			log.Printf("render template: %v", err)
		}
		return
	}
	var usr *model.User
	if u, ok := user.FromContext(r.Context()); ok {
		usr = &u
	}
	if err := views.Home(w, usr, nil); err != nil {
		log.Printf("render template: %v", err)
	}
}

func handleAccountPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	if err := views.Account(w, &u); err != nil {
		log.Printf("render template: %v", err)
	}
}

func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if err := views.Login(w); err != nil {
		log.Printf("render template: %v", err)
	}
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := user.FromContext(r.Context())
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
