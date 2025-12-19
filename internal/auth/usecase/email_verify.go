package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type EmailVerifyInput struct {
	Token string `validate:"required"`
}

func (s *Usecase) EmailVerify(ctx context.Context, in EmailVerifyInput) error {
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	// claims, err := s.jwt.Verify(in.Token)
	// if err != nil {
	// 	slog.WarnContext(ctx, "invalid email verify token", "error", err)
	// 	return goerror.NewBusiness("invalid verification token", goerror.CodeUnauthorized)
	// }

	// userID, err := strconv.ParseInt(claims.Subject(), 10, 64)
	// if err != nil {
	// 	slog.WarnContext(ctx, "invalid email verify subject", "subject", claims.Subject(), "error", err)
	// 	return goerror.NewBusiness("invalid verification token", goerror.CodeUnauthorized)
	// }

	// payload := claims.Payload()
	// if v, ok := payload["purpose"].(string); !ok || v != emailVerifyPurpose {
	// 	slog.WarnContext(ctx, "invalid email verify purpose", "user_id", userID)
	// 	return goerror.NewBusiness("invalid verification token", goerror.CodeUnauthorized)
	// }

	// user, err := s.repoDB.UserGetByID(ctx, userID)
	// if errors.Is(err, goerror.ErrNotFound) {
	// 	slog.WarnContext(ctx, "user not found for email verify", "user_id", userID)
	// 	return goerror.NewBusiness("invalid verification token", goerror.CodeUnauthorized)
	// }
	// if err != nil {
	// 	slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", userID, "error", err)
	// 	return goerror.NewServer(err)
	// }

	// switch user.Status {
	// case entity.UserStatusActive:
	// 	return nil

	// case entity.UserStatusBanned:
	// 	return goerror.NewBusiness("user account is banned", goerror.CodeForbidden)

	// case entity.UserStatusUnverified:
	// 	if err := s.repoDB.UserUpdateStatus(ctx, userID, entity.UserStatusUnverified, entity.UserStatusActive); err != nil {
	// 		slog.ErrorContext(ctx, "failed to repo update user status", "user_id", userID, "error", err)
	// 		return goerror.NewServer(err)
	// 	}

	// default:
	// 	slog.WarnContext(ctx, "unknown user status", "user_id", userID, "status", user.Status)
	// }

	return nil
}
