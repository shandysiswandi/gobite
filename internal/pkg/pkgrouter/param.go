package pkgrouter

import (
	"context"

	"github.com/julienschmidt/httprouter"
)

// GetParam reads a path parameter from the request context (as stored by httprouter).
func GetParam(ctx context.Context, key string) string {
	return httprouter.ParamsFromContext(ctx).ByName(key)
}
