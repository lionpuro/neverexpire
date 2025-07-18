package api

import (
	"context"

	"github.com/lionpuro/neverexpire/keys"
)

type ctxKey int

const (
	ctxKeyAPIKey ctxKey = iota
)

func ctxAPIKey(ctx context.Context) (keys.AccessKey, bool) {
	key, ok := ctx.Value(ctxKeyAPIKey).(keys.AccessKey)
	return key, ok
}
