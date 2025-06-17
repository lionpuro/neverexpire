package http

import (
	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/user"
)

type Handler struct {
	UserService   *user.Service
	DomainService *domain.Service
	AuthService   *auth.Service
}

func NewHandler(us *user.Service, ds *domain.Service, as *auth.Service) *Handler {
	return &Handler{
		UserService:   us,
		DomainService: ds,
		AuthService:   as,
	}
}
