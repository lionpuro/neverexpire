package http

import (
	"fmt"
	"net/http"

	"github.com/lionpuro/neverexpire/http/views"
)

func (h *Handler) APIPage(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	keys, err := h.KeyService.ByUser(r.Context(), u.ID)
	if err != nil {
		h.htmxError(w, fmt.Errorf("failed to load api keys"))
		h.log.Error("failed to load api keys", "error", err.Error())
		return
	}
	h.render(views.API(w, views.LayoutData{User: &u}, keys))
}

func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	raw, _, err := h.KeyService.Create(u.ID)
	if err != nil {
		h.htmxError(w, fmt.Errorf("failed to generate key"))
		return
	}
	h.render(views.Component(w, "api-key", map[string]string{"RawKey": raw}))
}

func (h *Handler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	id := r.PathValue("id")
	if id == "" {
		h.htmxError(w, fmt.Errorf("failed to delete token"))
		return
	}
	if err := h.KeyService.Delete(id, u.ID); err != nil {
		h.htmxError(w, fmt.Errorf("failed to delete token"))
		return
	}
	w.Header().Set("HX-Location", "/api")
	w.WriteHeader(http.StatusNoContent)
}
