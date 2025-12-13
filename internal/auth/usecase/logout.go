package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) Logout(ctx context.Context, in domain.LogoutInput) (*domain.LogoutOutput, error) {
	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	if err := s.validator.Validate(in); err != nil {
		return nil, err
	}

	refClaims, err := s.jwtRefreshToken.Verify(in.RefreshToken)
	if err != nil {
		slog.WarnContext(ctx, "invalid refresh token", "error", err)
		return nil, pkgerror.ErrAuthUnauthenticated
	}

	if clm.Subject() != refClaims.Subject() {
		slog.WarnContext(ctx, "token subject mismatch", "access_subject", clm.Subject(), "refresh_subject", refClaims.Subject())
		return nil, pkgerror.ErrAuthUnauthenticated
	}

	if err := s.repoCache.DeleteTokensID(ctx, clm.ID(), refClaims.ID()); err != nil {
		slog.ErrorContext(ctx, "failed to delete tokens jti from cache", "access_jti", clm.ID(), "refresh_jti", refClaims.ID(), "error", err)
		return nil, err
	}

	return &domain.LogoutOutput{Success: true}, nil
}
