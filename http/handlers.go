package http

import (
	"log"
	"net/http"

	"github.com/lionpuro/neverexpire/http/views"
	"github.com/lionpuro/neverexpire/model"
	"github.com/lionpuro/neverexpire/user"
)

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
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

func isHXrequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func htmxError(w http.ResponseWriter, err error) {
	w.Header().Set("HX-Retarget", "#banner-container")
	if err := views.ErrorBanner(w, err); err != nil {
		log.Printf("render error: %v", err)
	}
}

func handleErrorPage(w http.ResponseWriter, msg string, code int) {
	if err := views.Error(w, code, msg); err != nil {
		log.Printf("render template: %v", err)
	}
}
