package handlers

import "context"

type contextKey int

const (
	keyUserID contextKey = iota
)

func userIDToContext(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, keyUserID, uid)
}

func userIDFromContext(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(keyUserID).(string)
	return uid, ok
}
