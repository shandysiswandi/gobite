package pkgrouter

import (
	"context"

	"github.com/julienschmidt/httprouter"
)

func GetParam(ctx context.Context, key string) string {
	return httprouter.ParamsFromContext(ctx).ByName(key)
}
