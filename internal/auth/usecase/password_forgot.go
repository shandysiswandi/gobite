package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type ForgotPasswordInput struct {
	Email string `validate:"required,lowercase,email"`
}

func (s *Usecase) ForgotPassword(ctx context.Context, in ForgotPasswordInput) error {
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	// email := strings.TrimSpace(in.Email)
	// user, err := s.repoDB.GetUserByEmail(ctx, email)
	// if errors.Is(err, goerror.ErrNotFound) {
	// 	slog.WarnContext(ctx, "password reset requested for unavailable user", "email", email, "error", err)
	// 	return nil
	// }
	// if err != nil {
	// 	slog.WarnContext(ctx, "failed to repo get user by email", "email", email, "error", err)
	// 	return goerror.NewServer(err)
	// }

	// cToken := s.oid.Generate()
	// cTokenHash, err := s.hash.Hash(cToken)
	// if err != nil {
	// 	slog.ErrorContext(ctx, "failed to hash token", "error", err)
	// 	return goerror.NewServer(err)
	// }

	// expiresAt := s.clock.Now().Add(time.Duration(s.cfg.GetInt("modules.auth.password_reset_ttl")) * time.Hour)
	// if err := s.repoDB.CreateUserPasswordReset(ctx, user.ID, string(cTokenHash), expiresAt); err != nil {
	// 	slog.ErrorContext(ctx, "failed to repo create user password reset", "user_id", user.ID, "error", err)
	// 	return goerror.NewServer(err)
	// }

	// if err := s.repoMessaging.PublishUserForgotPassword(ctx, entity.UserForgotPasswordMessage{
	// 	Email:    user.Email,
	// 	FullName: user.FullName,
	// 	Token:    cToken,
	// }); err != nil {
	// 	slog.ErrorContext(ctx, "failed to publish user forgot password", "user_id", user.ID, "error", err)
	// }

	return nil
}
