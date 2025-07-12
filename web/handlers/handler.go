package handlers

import (
	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/users"
)

type Handler struct {
	UserService *users.Service
	HostService *hosts.Service
	AuthService *auth.Service
	KeyService  *keys.Service
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
		UserService: us,
		HostService: hs,
		AuthService: as,
		KeyService:  ks,
		log:         logger,
	}
}

func (h *Handler) render(err error) {
	if err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}
