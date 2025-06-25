package http

import (
	"net/http"

	"github.com/lionpuro/neverexpire/http/views"
	"github.com/lionpuro/neverexpire/model"
)

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		if err := views.Error(w, http.StatusNotFound, "Page not found"); err != nil {
			h.log.Error("failed to render template", "error", err.Error())
		}
		return
	}
	var usr *model.User
	if u, ok := userFromContext(r.Context()); ok {
		usr = &u
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

func (h *Handler) ErrorPage(w http.ResponseWriter, msg string, code int) {
	if err := views.Error(w, code, msg); err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}
