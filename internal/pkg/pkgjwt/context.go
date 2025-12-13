package pkgjwt

import "context"

type jwtContextKey struct{}

func GetAuth[T any](ctx context.Context) *Claims[T] {
	clm, ok := ctx.Value(jwtContextKey{}).(*Claims[T])
	if !ok {
		return nil
	}

	return clm
}

func SetAuth[T any](ctx context.Context, clm Claims[T]) context.Context {
	return context.WithValue(ctx, jwtContextKey{}, clm)
}
