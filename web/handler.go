package web

import (
	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/users"
)

type Handler struct {
	userService   *users.Service
	hostService   *hosts.Service
	keyService    *keys.Service
	Authenticator *auth.Authenticator
	log           logging.Logger
}

func NewHandler(
	logger logging.Logger,
	us *users.Service,
	hs *hosts.Service,
	ks *keys.Service,
	auth *auth.Authenticator,
) *Handler {
	return &Handler{
		userService:   us,
		hostService:   hs,
		keyService:    ks,
		Authenticator: auth,
		log:           logger,
	}
}

func (h *Handler) render(err error) {
	if err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}
