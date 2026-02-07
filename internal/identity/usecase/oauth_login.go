package usecase

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/identity/outbound/oauth"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type OAuthStartInput struct {
	Provider     string `validate:"required"`
	RedirectPath string
}

type OAuthStartOutput struct {
	AuthURL string
}

type OAuthCallbackInput struct {
	Provider string `validate:"required"`
	Code     string `validate:"required"`
	State    string `validate:"required"`
}

type OAuthCallbackOutput struct {
	MfaRequired      bool
	ChallengeToken   string
	AvailableMethods []string
	AccessToken      string
	RefreshToken     string
	RedirectPath     string
}

func (s *Usecase) OAuthStart(ctx context.Context, in OAuthStartInput) (*OAuthStartOutput, error) {
	ctx, span := s.startSpan(ctx, "OAuthStart")
	defer span.End()

	resp, err := s.oauth.Start(ctx, oauth.StartInput{
		Provider:     in.Provider,
		RedirectPath: in.RedirectPath,
	})
	if err != nil {
		return nil, err
	}

	return &OAuthStartOutput{AuthURL: resp.AuthURL}, nil
}

func (s *Usecase) OAuthCallback(ctx context.Context, in OAuthCallbackInput) (*OAuthCallbackOutput, error) {
	ctx, span := s.startSpan(ctx, "OAuthCallback")
	defer span.End()

	oauthResp, err := s.oauth.Callback(ctx, oauth.CallbackInput{
		Provider: in.Provider,
		Code:     in.Code,
		State:    in.State,
	})
	if err != nil {
		redirectPath := oauth.DefaultRedirectPath
		if oauthResp != nil && strings.TrimSpace(oauthResp.RedirectPath) != "" {
			redirectPath = oauthResp.RedirectPath
		}
		return &OAuthCallbackOutput{RedirectPath: redirectPath}, err
	}

	user, err := s.resolveOAuthUser(ctx, oauthResp.Provider, oauthResp.Profile)
	if err != nil {
		return &OAuthCallbackOutput{RedirectPath: oauthResp.RedirectPath}, err
	}

	if err := s.ensureUserStatusAllowed(ctx, user.ID, user.Status); err != nil {
		return &OAuthCallbackOutput{RedirectPath: oauthResp.RedirectPath}, err
	}

	output, err := s.issueOAuthTokens(ctx, user)
	if err != nil {
		return &OAuthCallbackOutput{RedirectPath: oauthResp.RedirectPath}, err
	}
	output.RedirectPath = oauthResp.RedirectPath
	return output, nil
}

func (s *Usecase) resolveOAuthUser(ctx context.Context, provider string, profile oauth.Profile) (*entity.User, error) {
	conn, err := s.repoDB.GetUserConnectionByProviderUserID(ctx, provider, profile.ProviderUserID)
	if err == nil {
		user, err := s.repoDB.GetUserByID(ctx, conn.UserID, false)
		if err != nil {
			slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", conn.UserID, "error", err)
			return nil, goerror.NewServer(err)
		}
		return user, nil
	}
	if !errors.Is(err, goerror.ErrNotFound) {
		slog.ErrorContext(ctx, "failed to repo get user connection", "provider", provider, "error", err)
		return nil, goerror.NewServer(err)
	}

	user, err := s.repoDB.GetUserByEmail(ctx, profile.Email, false)
	if err == nil {
		if user.Status == entity.UserStatusUnverified {
			if err := s.repoDB.PatchUser(ctx, entity.PatchUser{
				ID:        user.ID,
				Status:    entity.UserStatusActive,
				UpdatedBy: user.ID,
			}, ""); err != nil {
				slog.ErrorContext(ctx, "failed to update user status", "user_id", user.ID, "error", err)
				return nil, goerror.NewServer(err)
			}
			user.Status = entity.UserStatusActive
		}

		if err := s.repoDB.CreateUserConnection(ctx, entity.UserConnection{
			ID:             s.uid.Generate(),
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: profile.ProviderUserID,
		}); err != nil {
			slog.ErrorContext(ctx, "failed to create user connection", "user_id", user.ID, "provider", provider, "error", err)
			return nil, goerror.NewServer(err)
		}

		return user, nil
	}
	if !errors.Is(err, goerror.ErrNotFound) {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", profile.Email, "error", err)
		return nil, goerror.NewServer(err)
	}

	fullName := strings.TrimSpace(profile.FullName)
	if fullName == "" {
		fullName = strings.TrimSpace(profile.Nickname)
	}
	if fullName == "" && profile.Email != "" {
		fullName = strings.Split(profile.Email, "@")[0]
	}
	if fullName == "" {
		fullName = "User"
	}

	avatarURL := strings.TrimSpace(profile.AvatarURL)
	if avatarURL == "" {
		avatarURL = "https://ui-avatars.com/api/?name=" + url.QueryEscape(fullName)
	}

	password, err := oauth.GenerateRandomToken(32)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate oauth password", "error", err)
		return nil, goerror.NewServer(err)
	}

	passwordHash, err := s.bcrypt.Hash(password)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash oauth password", "error", err)
		return nil, goerror.NewServer(err)
	}

	newUserID := s.uid.Generate()
	newUser := entity.NewUser{
		ID:        newUserID,
		Email:     profile.Email,
		FullName:  fullName,
		AvatarURL: avatarURL,
		Status:    entity.UserStatusActive,
		CreatedBy: newUserID,
		UpdatedBy: newUserID,
	}
	newConn := entity.UserConnection{
		ID:             s.uid.Generate(),
		UserID:         newUserID,
		Provider:       provider,
		ProviderUserID: profile.ProviderUserID,
	}

	if err := s.repoDB.NewOAuthUser(ctx, newUser, string(passwordHash), newConn); err != nil {
		slog.ErrorContext(ctx, "failed to repo create oauth user", "email", newUser.Email, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &entity.User{
		ID:        newUser.ID,
		Email:     newUser.Email,
		FullName:  newUser.FullName,
		AvatarURL: newUser.AvatarURL,
		Status:    newUser.Status,
	}, nil
}

func (s *Usecase) issueOAuthTokens(ctx context.Context, user *entity.User) (*OAuthCallbackOutput, error) {
	mfaFactors, err := s.repoDB.GetMFAFactorByUserID(ctx, user.ID, true)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get mfa factors", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	if len(mfaFactors) > 0 {
		cToken := s.oid.Generate()
		cTokenHash, err := s.hmac.Hash(cToken)
		if err != nil {
			slog.ErrorContext(ctx, "failed to hash token challenge", "error", err)
			return nil, goerror.NewServer(err)
		}

		if err := s.repoDB.CreateChallenge(ctx, entity.Challenge{
			ID:        s.uid.Generate(),
			UserID:    user.ID,
			Token:     string(cTokenHash),
			Purpose:   entity.ChallengePurposeMFALogin,
			ExpiresAt: s.clock.Now().Add(s.cfg.GetMinute("modules.identity.mfa_login_ttl_minutes")),
		}); err != nil {
			slog.ErrorContext(ctx, "failed to repo create challenge", "user_id", user.ID, "error", err)
			return nil, goerror.NewServer(err)
		}

		return &OAuthCallbackOutput{
			MfaRequired:      true,
			ChallengeToken:   cToken,
			AvailableMethods: []string{entity.MFATypeTOTP.String(), entity.MFATypeBackupCode.String()},
		}, nil
	}

	acToken, err := s.jwt.Generate(user.ID, user.Email)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate access jwt token", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	refToken := s.oid.Generate()
	refTokenHash, err := s.hmac.Hash(refToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash refresh token", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	if err := s.repoDB.CreateRefreshToken(ctx, entity.RefreshToken{
		ID:        s.uid.Generate(),
		UserID:    user.ID,
		Token:     string(refTokenHash),
		ExpiresAt: s.clock.Now().Add(s.cfg.GetDay("modules.identity.refresh_token_ttl_days")),
	}); err != nil {
		slog.ErrorContext(ctx, "failed to repo create refresh token user", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &OAuthCallbackOutput{
		AccessToken:  acToken,
		RefreshToken: refToken,
	}, nil
}
