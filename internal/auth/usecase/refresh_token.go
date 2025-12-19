package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
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
		return nil, goerror.NewInvalidInput(err)
	}

	oldRefreshTokenHash, err := s.hash.Hash(in.RefreshToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash old refresh token", "error", err)
		return nil, goerror.NewServer(err)
	}

	rt, err := s.repoDB.GetUserRefreshToken(ctx, string(oldRefreshTokenHash))
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user refresh token not found")
		return nil, goerror.NewBusiness("invalid or expired refresh token", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user refresh token", "error", err)
		return nil, goerror.NewServer(err)
	}

	// SECURITY CHECK: Reuse Detection for rotated tokens only.
	if rt.RefreshRevoked {
		if rt.RefreshReplacedByTokenID != nil {
			// CRITICAL: The user is trying to use a token that was already rotated.
			// This implies the token was stolen. Invalidate ALL tokens for this user.
			if err := s.repoDB.RevokeAllRefreshToken(ctx, rt.UserID); err != nil {
				slog.ErrorContext(ctx, "failed to repo revoke all refresh token", "user_id", rt.UserID, "error", err)
			}

			slog.WarnContext(ctx, "SECURITY: refresh token reuse detected")
			return nil, goerror.NewBusiness("token reuse detected, please log in again", goerror.CodeForbidden)
		}

		slog.WarnContext(ctx, "refresh token is revoked", "refresh_token_id", rt.RefreshID)
		return nil, goerror.NewBusiness("invalid or expired refresh token", goerror.CodeUnauthorized)
	}

	if s.clock.Now().After(rt.RefreshExpiresAt) {
		slog.WarnContext(ctx, "user refresh token is expired")
		return nil, goerror.NewBusiness("invalid or expired refresh token", goerror.CodeUnauthorized)
	}

	if err := s.ensureUserAllowedToLogin(ctx, rt.UserID, rt.UserStatus); err != nil {
		return nil, err
	}

	newRefreshToken := s.oid.Generate()
	newRefreshTokenHash, err := s.hash.Hash(newRefreshToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash new refresh token", "error", err)
		return nil, goerror.NewServer(err)
	}

	acToken, err := s.jwt.Generate(
		jwt.WithID(s.uuid.Generate()),
		jwt.WithSubject(strconv.FormatInt(rt.UserID, 10)),
		jwt.WithPayloadValue(keyPayloadUserID, rt.UserID),
		jwt.WithPayloadValue(keyPayloadUserEmail, rt.UserEmail),
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate access jwt token", "user_id", rt.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	err = s.repoDB.RotateRefreshToken(ctx, entity.RotateRefreshToken{
		NewID:        s.uid.Generate(),
		OldID:        rt.RefreshID,
		UserID:       rt.UserID,
		NewToken:     string(newRefreshTokenHash),
		NewExpiresAt: s.clock.Now().Add(time.Duration(s.cfg.GetInt("modules.auth.refresh_token_ttl")) * 24 * time.Hour),
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo rotate refresh token", "error", err)
		return nil, goerror.NewServer(err)
	}

	return &RefreshTokenOutput{
		AccessToken:  acToken,
		RefreshToken: newRefreshToken,
	}, nil
}
