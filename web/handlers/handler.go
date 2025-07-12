package handlers

import (
	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/users"
)

type Handler struct {
	userService *users.Service
	hostService *hosts.Service
	AuthService *auth.Service
	keyService  *keys.Service
	log         logging.Logger
}

func New(
	logger logging.Logger,
	us *users.Service,
	hs *hosts.Service,
	ks *keys.Service,
	as *auth.Service,
) *Handler {
	return &Handler{
		userService: us,
		hostService: hs,
		AuthService: as,
		keyService:  ks,
		log:         logger,
	}
}

func (h *Handler) render(err error) {
	if err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}
