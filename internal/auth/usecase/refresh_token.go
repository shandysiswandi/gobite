package usecase

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

func (s *Usecase) RefreshToken(ctx context.Context, in domain.RefreshTokenInput) (*domain.RefreshTokenOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, err
	}

	refClaims, err := s.jwtRefreshToken.Verify(in.RefreshToken)
	if err != nil {
		slog.WarnContext(ctx, "invalid refresh token", "error", err)
		return nil, pkgerror.ErrUnauthenticated
	}

	isStillValid, err := s.repoCache.IsTokenIDExist(ctx, refClaims.ID())
	if err != nil {
		slog.ErrorContext(ctx, "failed to check refresh token jti", "refresh_jti", refClaims.ID(), "error", err)
		return nil, err
	}
	if !isStillValid {
		slog.WarnContext(ctx, "refresh token revoked or expired in cache", "refresh_jti", refClaims.ID())
		return nil, pkgerror.ErrUnauthenticated
	}

	userID, err := strconv.ParseInt(refClaims.Subject(), 10, 64)
	if err != nil {
		slog.ErrorContext(ctx, "invalid refresh token subject", "subject", refClaims.Subject(), "error", err)
		return nil, pkgerror.ErrUnauthenticated
	}

	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	acToken, acJTI, refToken, refJTI, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := s.repoCache.RotateTokensID(ctx, refClaims.ID(), acJTI, refJTI); err != nil {
		slog.ErrorContext(ctx, "failed to rotate tokens jti", "old_refresh_jti", refClaims.ID(), "new_access_jti", acJTI, "new_refresh_jti", refJTI, "error", err)
		return nil, err
	}

	return &domain.RefreshTokenOutput{
		AccessToken:  acToken,
		RefreshToken: refToken,
	}, nil
}
