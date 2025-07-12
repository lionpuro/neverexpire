package web

import (
	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/user"
	"github.com/lionpuro/neverexpire/web/handlers"
)

func NewHandler(
	logger logging.Logger,
	u *user.Service,
	h *hosts.Service,
	k *keys.Service,
	a *auth.Service,
) *handlers.Handler {
	return handlers.New(logger, u, h, k, a)
}
