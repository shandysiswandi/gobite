package usecase

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

type LoginInput struct {
	Email    string `validate:"required,lowercase,email"`
	Password string `validate:"required"`
}

type LoginOutput struct {
	MfaRequired  bool
	PreAuthToken string
	//
	AccessToken  string
	RefreshToken string
}

func (s *Usecase) Login(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, pkgerror.NewInvalidInput(err)
	}

	user, err := s.getUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}

	userCred, err := s.getCredential(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if !s.hash.Verify(userCred.Password, in.Password) {
		slog.WarnContext(ctx, "password user account not match")
		return nil, pkgerror.NewBusiness("invalid email or password", pkgerror.CodeUnauthorized)
	}

	mfaFacs, err := s.repoDB.MfaFactorGetByUserID(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get mfa factor by user_id", "user_id", user.ID, "error", err)
		return nil, pkgerror.NewServer(err)
	}

	// this mean user has mfa active
	if len(mfaFacs) > 0 {
		strID := strconv.FormatInt(user.ID, 10)
		tempToken, _, err := s.jwtTempToken.Generate(strID, map[string]any{"some_id": mfaFacs[0].ID})
		if err != nil {
			slog.ErrorContext(ctx, "failed to generate temp jwt token", "user_id", user.ID, "error", err)
			return nil, pkgerror.NewServer(err)
		}

		return &LoginOutput{
			MfaRequired:  true,
			PreAuthToken: tempToken,
		}, nil
	}

	acToken, acJTI, refToken, refJTI, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := s.repoCache.SaveTokensID(ctx, acJTI, refJTI); err != nil {
		slog.ErrorContext(ctx, "failed to save jtis to redis", "ac", acJTI, "ref", refJTI, "error", err)
		return nil, pkgerror.NewServer(err)
	}

	return &LoginOutput{
		AccessToken:  acToken,
		RefreshToken: refToken,
	}, nil
}
