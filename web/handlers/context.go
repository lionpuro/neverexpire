package handlers

import (
	"context"

	"github.com/lionpuro/neverexpire/user"
)

type contextKey int

const (
	userKey contextKey = iota
)

func userToContext(ctx context.Context, user user.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func userFromContext(ctx context.Context) (user.User, bool) {
	u, ok := ctx.Value(userKey).(user.User)
	return u, ok
}
