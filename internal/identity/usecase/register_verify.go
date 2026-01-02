package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type RegisterVerifyInput struct {
	ChallengeToken string `validate:"required"`
}

func (s *Usecase) RegisterVerify(ctx context.Context, in RegisterVerifyInput) error {
	ctx, span := s.startSpan(ctx, "RegisterVerify")
	defer span.End()

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	cTokenHash, err := s.hmac.Hash(in.ChallengeToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash token challange", "error", err)
		return goerror.NewServer(err)
	}

	cu, err := s.repoDB.GetChallengeUserByTokenPurpose(ctx, string(cTokenHash), entity.ChallengePurposeRegisterVerify)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "challenge user not found", "challenge_token", string(cTokenHash))
		return goerror.NewBusiness("invalid verification token", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get challange user by token purpose", "challenge_token", string(cTokenHash), "error", err)
		return goerror.NewServer(err)
	}

	switch cu.UserStatus.Ensure() {
	case entity.UserStatusActive:
		if err := s.repoDB.DeleteChallenge(ctx, cu.ChallengeID); err != nil {
			slog.ErrorContext(ctx, "failed to repo delete challenge by id", "challenge_id", cu.ChallengeID, "error", err)
			return goerror.NewServer(err)
		}

		return nil

	case entity.UserStatusBanned:
		return goerror.NewBusiness("user account is banned", goerror.CodeForbidden)

	case entity.UserStatusUnverified:
		if err := s.repoDB.VerifyUserRegistration(ctx, entity.VerifyUserRegistration{
			ChallengeID:   cu.ChallengeID,
			UserID:        cu.UserID,
			UpdatedBy:     cu.UserID,
			OldUserStatus: entity.UserStatusUnverified,
			NewUserStatus: entity.UserStatusActive,
		}); err != nil {
			slog.ErrorContext(ctx, "failed to repo verify user registration", "user_id", cu.UserID, "challenge_id", cu.ChallengeID, "error", err)
			return goerror.NewServer(err)
		}

		return nil

	default:
		slog.WarnContext(ctx, "unknown user status", "user_id", cu.UserID, "status", cu.UserStatus.String())
		return goerror.NewBusiness("account status is unrecognized", goerror.CodeForbidden)
	}
}
