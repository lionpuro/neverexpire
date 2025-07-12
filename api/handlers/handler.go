package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/user"
)

type Handler struct {
	userService   *user.Service
	domainService *domain.Service
	keyService    *keys.Service
	log           logging.Logger
}

func New(
	logger logging.Logger,
	us *user.Service,
	ds *domain.Service,
	ks *keys.Service,
) *Handler {
	return &Handler{
		userService:   us,
		domainService: ds,
		keyService:    ks,
		log:           logger,
	}
}

func (h *Handler) json(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.log.Error("failed to write json response: %v", err)
	}
}
