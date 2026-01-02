package instrument

import "context"

type chainIDContextKey struct{}

// GetCorrelationID returns the correlation ID stored in the context.
//
// Middleware is expected to set this value early in the request lifecycle so
// it can be attached to logs and propagated to downstream calls.
func GetCorrelationID(ctx context.Context) string {
	clm, ok := ctx.Value(chainIDContextKey{}).(string)
	if !ok {
		return "[invalid_chain_id]"
	}
	return clm
}

// SetCorrelationID stores a correlation ID into the context.
func SetCorrelationID(ctx context.Context, cid string) context.Context {
	return context.WithValue(ctx, chainIDContextKey{}, cid)
}
