package web

import (
	"net/http"
	"strings"

	"github.com/lionpuro/neverexpire/users"
	"github.com/lionpuro/neverexpire/web/views"
)

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	var usr *users.User
	if u, ok := userFromContext(r.Context()); ok {
		usr = &u
	}
	if r.URL.Path != "/" {
		if strings.HasPrefix(r.URL.Path, "/api") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			if _, err := w.Write([]byte("Not found")); err != nil {
				h.log.Error("failed to write response", "error", err.Error())
			}
			return
		}
		h.render(views.Error(w, views.LayoutData{User: usr}, http.StatusNotFound, "Page not found"))
		return
	}
	h.render(views.Home(w, views.LayoutData{User: usr}))
}

func isHXrequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func (h *Handler) htmxError(w http.ResponseWriter, err error) {
	w.Header().Set("HX-Retarget", "#banner-container")
	h.render(views.ErrorBanner(w, err))
}

func (h *Handler) ErrorPage(w http.ResponseWriter, r *http.Request, msg string, code int) {
	var usr *users.User
	if u, ok := userFromContext(r.Context()); ok {
		usr = &u
	}
	h.render(views.Error(w, views.LayoutData{User: usr}, code, msg))
}

func (h *Handler) PrivacyPage(w http.ResponseWriter, r *http.Request) {
	var usr *users.User
	if u, ok := userFromContext(r.Context()); ok {
		usr = &u
	}
	h.render(views.Privacy(w, views.LayoutData{User: usr}))
}
