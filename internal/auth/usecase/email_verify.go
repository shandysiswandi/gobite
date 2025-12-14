package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

const emailVerifyPurpose = "email_verify"

func (s *Usecase) EmailVerify(ctx context.Context, in domain.EmailVerifyInput) error {
	if err := s.validator.Validate(in); err != nil {
		return pkgerror.NewInvalidInput(err)
	}

	claims, err := s.jwtTempToken.Verify(in.Token)
	if err != nil {
		slog.WarnContext(ctx, "invalid email verify token", "error", err)
		return pkgerror.NewBusiness("invalid verification token", pkgerror.CodeUnauthorized)
	}

	userID, err := strconv.ParseInt(claims.Subject(), 10, 64)
	if err != nil {
		slog.WarnContext(ctx, "invalid email verify subject", "subject", claims.Subject(), "error", err)
		return pkgerror.NewBusiness("invalid verification token", pkgerror.CodeUnauthorized)
	}

	payload := claims.Payload()
	if v, ok := payload["purpose"].(string); !ok || v != emailVerifyPurpose {
		slog.WarnContext(ctx, "invalid email verify purpose", "user_id", userID)
		return pkgerror.NewBusiness("invalid verification token", pkgerror.CodeUnauthorized)
	}

	user, err := s.repoDB.UserGetByID(ctx, userID)
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user not found for email verify", "user_id", userID)
		return pkgerror.NewBusiness("invalid verification token", pkgerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", userID, "error", err)
		return pkgerror.NewServer(err)
	}

	switch user.Status {
	case domain.UserStatusActive:
		return nil

	case domain.UserStatusBanned:
		return pkgerror.NewBusiness("user account is banned", pkgerror.CodeForbidden)

	case domain.UserStatusUnverified:
		if err := s.repoDB.UserUpdateStatus(ctx, userID, domain.UserStatusUnverified, domain.UserStatusActive); err != nil {
			slog.ErrorContext(ctx, "failed to repo update user status", "user_id", userID, "error", err)
			return pkgerror.NewServer(err)
		}

	default:
		slog.WarnContext(ctx, "unknown user status", "user_id", userID, "status", user.Status)
	}

	return nil
}
