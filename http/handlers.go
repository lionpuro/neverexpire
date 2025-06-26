package http

import (
	"net/http"

	"github.com/lionpuro/neverexpire/http/views"
	"github.com/lionpuro/neverexpire/model"
)

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	var usr *model.User
	if u, ok := userFromContext(r.Context()); ok {
		usr = &u
	}
	if r.URL.Path != "/" {
		if err := views.Error(w, usr, http.StatusNotFound, "Page not found"); err != nil {
			h.log.Error("failed to render template", "error", err.Error())
		}
		return
	}
	if err := views.Home(w, usr, nil); err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}

func isHXrequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func (h *Handler) htmxError(w http.ResponseWriter, err error) {
	w.Header().Set("HX-Retarget", "#banner-container")
	if err := views.ErrorBanner(w, err); err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}

func (h *Handler) ErrorPage(w http.ResponseWriter, r *http.Request, msg string, code int) {
	var usr *model.User
	if u, ok := userFromContext(r.Context()); ok {
		usr = &u
	}
	if err := views.Error(w, usr, code, msg); err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}
