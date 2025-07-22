package web

import (
	"context"

	"github.com/lionpuro/neverexpire/users"
)

type contextKey int

const (
	userKey contextKey = iota
)

func userToContext(ctx context.Context, user users.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func userFromContext(ctx context.Context) (users.User, bool) {
	u, ok := ctx.Value(userKey).(users.User)
	return u, ok
}
