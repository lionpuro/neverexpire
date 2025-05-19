package user

import (
	"context"

	"github.com/lionpuro/trackcerts/model"
)

const (
	userContextKey = "user"
)

func SaveToContext(ctx context.Context, user model.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func FromContext(ctx context.Context) (model.User, bool) {
	u, ok := ctx.Value(userContextKey).(model.User)
	return u, ok
}
