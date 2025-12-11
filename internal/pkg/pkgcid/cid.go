package pkgcid

import "context"

type chainIDContextKey struct{}

func GetCorrelationID(ctx context.Context) string {
	clm, ok := ctx.Value(chainIDContextKey{}).(string)
	if !ok {
		return "[invalid_chain_id]"
	}
	return clm
}

func SetCorrelationID(ctx context.Context, cid string) context.Context {
	return context.WithValue(ctx, chainIDContextKey{}, cid)
}
