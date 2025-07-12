package handlers

import (
	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/user"
)

type Handler struct {
	UserService   *user.Service
	DomainService *domain.Service
	AuthService   *auth.Service
	KeyService    *keys.Service
	log           logging.Logger
}

func New(
	logger logging.Logger,
	us *user.Service,
	ds *domain.Service,
	ks *keys.Service,
	as *auth.Service,
) *Handler {
	return &Handler{
		UserService:   us,
		DomainService: ds,
		AuthService:   as,
		KeyService:    ks,
		log:           logger,
	}
}

func (h *Handler) render(err error) {
	if err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}
