package api

import (
	"github.com/lionpuro/neverexpire/api/handlers"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/users"
)

func NewHandler(
	logger logging.Logger,
	us *users.Service,
	hs *hosts.Service,
	ks *keys.Service,
) *handlers.Handler {
	return handlers.New(logger, us, hs, ks)
}
