package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/users"
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

func (h *Handler) CreateHost(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r.Context())
	var body struct {
		Hostname string `json:"hostname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.json(w, http.StatusBadRequest, "Bad request")
		return
	}
	s := strings.TrimSpace(body.Hostname)
	if s == "" {
		h.json(w, http.StatusBadRequest, "Bad request")
		return
	}
	name, err := hosts.ParseHostname(body.Hostname)
	if err != nil {
		h.json(w, http.StatusBadRequest, "Bad request")
		return
	}
	if err := h.hostService.Create(users.User{ID: uid}, []string{name}); err != nil {
		if strings.Contains(err.Error(), "already tracking") {
			h.json(w, http.StatusBadRequest, "The host is already being tracked")
			return
		}
		h.json(w, http.StatusBadRequest, "Bad request")
		return
	}
	host, err := h.hostService.ByName(r.Context(), name, uid)
	if err != nil {
		h.log.Error("failed to retrieve new host by name", "error", err.Error())
		h.json(w, http.StatusInternalServerError, "Error retrieving created host")
		return
	}
	data := map[string]interface{}{"data": host}
	h.json(w, http.StatusCreated, data)
}

func (h *Handler) DeleteHost(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r.Context())
	name := r.PathValue("hostname")
	host, err := h.hostService.ByName(r.Context(), name, uid)
	if err != nil {
		if db.IsErrNoRows(err) {
			h.json(w, http.StatusNotFound, "Host not found")
			return
		}
		h.log.Error("failed to get host", "error", err.Error())
		h.json(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if err := h.hostService.Delete(uid, host.ID); err != nil {
		h.log.Error("failed to delete host", "error", err.Error())
		h.json(w, http.StatusInternalServerError, "failed to delete host")
		return
	}
	h.json(w, http.StatusAccepted, "Deleted successfully")
}
