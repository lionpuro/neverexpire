package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/users"
)

type Handler struct {
	userService *users.Service
	hostService *hosts.Service
	keyService  *keys.Service
	log         logging.Logger
}

func New(
	logger logging.Logger,
	us *users.Service,
	hs *hosts.Service,
	ks *keys.Service,
) *Handler {
	return &Handler{
		userService: us,
		hostService: hs,
		keyService:  ks,
		log:         logger,
	}
}

func (h *Handler) json(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.log.Error("failed to write json response: %v", err)
	}
}
