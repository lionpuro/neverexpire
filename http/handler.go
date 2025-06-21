package http

import (
	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/user"
)

type Handler struct {
	UserService   *user.Service
	DomainService *domain.Service
	AuthService   *auth.Service
	log           logging.Logger
}

func NewHandler(logger logging.Logger, us *user.Service, ds *domain.Service, as *auth.Service) *Handler {
	return &Handler{
		UserService:   us,
		DomainService: ds,
		AuthService:   as,
		log:           logger,
	}
}
