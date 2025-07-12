package api

import (
	"github.com/lionpuro/neverexpire/api/handlers"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/user"
)

func NewHandler(
	logger logging.Logger,
	us *user.Service,
	hs *hosts.Service,
	ks *keys.Service,
) *handlers.Handler {
	return handlers.New(logger, us, hs, ks)
}
