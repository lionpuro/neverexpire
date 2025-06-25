package http

import (
	"context"

	"github.com/lionpuro/neverexpire/model"
)

type contextKey int

const (
	userKey contextKey = iota
)

func userToContext(ctx context.Context, user model.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func userFromContext(ctx context.Context) (model.User, bool) {
	u, ok := ctx.Value(userKey).(model.User)
	return u, ok
}
