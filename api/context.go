package api

import "context"

type ctxKey int

const (
	ctxKeyUID ctxKey = iota
)

func currentUID(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(ctxKeyUID).(string)
	return uid, ok
}
