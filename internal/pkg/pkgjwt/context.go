package pkgjwt

import "context"

type jwtContextKey struct{}

// GetAuth returns the JWT claims stored in the context, if any.
func GetAuth[T any](ctx context.Context) *Claims[T] {
	clm, ok := ctx.Value(jwtContextKey{}).(Claims[T])
	if !ok {
		return nil
	}

	return &clm
}

// SetAuth stores JWT claims in the context.
func SetAuth[T any](ctx context.Context, clm Claims[T]) context.Context {
	return context.WithValue(ctx, jwtContextKey{}, clm)
}
