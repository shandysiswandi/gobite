package jwt

import "context"

type jwtContextKey struct{}

// GetAuth returns the JWT claims stored in the context, if any.
func GetAuth(ctx context.Context) *Claims {
	clm, ok := ctx.Value(jwtContextKey{}).(Claims)
	if !ok {
		return nil
	}

	return &clm
}

// SetAuth stores JWT claims in the context.
func SetAuth(ctx context.Context, clm Claims) context.Context {
	return context.WithValue(ctx, jwtContextKey{}, clm)
}
