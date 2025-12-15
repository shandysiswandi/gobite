package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

type RefreshTokenInput struct {
	RefreshToken string `validate:"required"`
}

type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string
}

func (s *Usecase) RefreshToken(ctx context.Context, in RefreshTokenInput) (*RefreshTokenOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, pkgerror.NewInvalidInput(err)
	}

	refClaims, err := s.jwtRefreshToken.Verify(in.RefreshToken)
	if err != nil {
		slog.WarnContext(ctx, "invalid refresh token", "error", err)
		return nil, pkgerror.NewBusiness("invalid refresh token", pkgerror.CodeUnauthorized)
	}

	isStillValid, err := s.repoCache.IsTokenIDExist(ctx, refClaims.ID())
	if err != nil {
		slog.ErrorContext(ctx, "failed to check refresh token jti", "refresh_jti", refClaims.ID(), "error", err)
		return nil, pkgerror.NewServer(err)
	}
	if !isStillValid {
		slog.WarnContext(ctx, "refresh token revoked or expired in cache", "refresh_jti", refClaims.ID())
		return nil, pkgerror.NewBusiness("refresh token revoked or expired", pkgerror.CodeUnauthorized)
	}

	user, err := s.getUserByID(ctx, refClaims.Payload().UserID)
	if err != nil {
		return nil, err
	}

	acToken, acJTI, refToken, refJTI, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := s.repoCache.RotateTokensID(ctx, refClaims.ID(), acJTI, refJTI); err != nil {
		slog.ErrorContext(ctx, "failed to rotate tokens jti", "old_refresh_jti", refClaims.ID(), "new_access_jti", acJTI, "new_refresh_jti", refJTI, "error", err)
		return nil, pkgerror.NewServer(err)
	}

	return &RefreshTokenOutput{
		AccessToken:  acToken,
		RefreshToken: refToken,
	}, nil
}
