package handlers

import (
	"net/http"

	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/hosts"
)

func (h *Handler) ListHosts(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r.Context())
	hsts, err := h.hostService.AllByUser(r.Context(), uid)
	if err != nil {
		h.log.Error("failed to get hosts", "error", err.Error())
		h.json(w, http.StatusInternalServerError, "internal server error")
		return
	}
	var result []hosts.APIModel
	for _, d := range hsts {
		host := hosts.ToAPIModel(d)
		result = append(result, host)
	}
	data := map[string]interface{}{
		"data": result,
	}
	h.json(w, http.StatusOK, data)
}

func (h *Handler) FindHost(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r.Context())
	name := r.PathValue("hostname")
	if name == "" {
		h.json(w, http.StatusBadRequest, "invalid hostname")
		return
	}
	host, err := h.hostService.ByName(r.Context(), name, uid)
	if err != nil {
		if db.IsErrNoRows(err) {
			h.json(w, http.StatusNotFound, "Not found")
			return
		}
		h.log.Error("failed to get host", "error", err.Error())
		h.json(w, http.StatusInternalServerError, "internal server error")
		return
	}
	result := hosts.ToAPIModel(host)
	data := map[string]interface{}{
		"data": result,
	}
	h.json(w, http.StatusOK, data)
}
