package http

import (
	"net/http"

	"github.com/lionpuro/neverexpire/http/views"
	"github.com/lionpuro/neverexpire/user"
)

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	var usr *user.User
	if u, ok := userFromContext(r.Context()); ok {
		usr = &u
	}
	if r.URL.Path != "/" {
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
	var usr *user.User
	if u, ok := userFromContext(r.Context()); ok {
		usr = &u
	}
	h.render(views.Error(w, views.LayoutData{User: usr}, code, msg))
}
