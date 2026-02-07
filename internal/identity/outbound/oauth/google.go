package oauth

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type GoogleProvider struct{}

func (GoogleProvider) Name() string {
	return "google"
}

func (GoogleProvider) ApplyDefaults(cfg ProviderConfig) ProviderConfig {
	cfg.RequireEmail = true
	cfg.RequireVerified = true
	if cfg.AuthURL == "" {
		cfg.AuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = "https://oauth2.googleapis.com/token"
	}
	if cfg.UserInfoURL == "" {
		cfg.UserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"openid", "email", "profile"}
	}
	return cfg
}

func (GoogleProvider) FetchProfile(ctx context.Context, client *http.Client, cfg ProviderConfig, accessToken string) (Profile, error) {
	profile := Profile{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserInfoURL, nil)
	if err != nil {
		return profile, goerror.NewServer(err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return profile, goerror.NewServer(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return profile, goerror.NewServer(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.WarnContext(ctx, "oauth userinfo failed", "status", resp.StatusCode, "body", string(body))
		return profile, goerror.NewBusiness("oauth userinfo failed", goerror.CodeUnauthorized)
	}

	var raw struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return profile, goerror.NewServer(err)
	}

	profile.ProviderUserID = raw.Sub
	profile.Email = raw.Email
	profile.EmailVerified = raw.EmailVerified
	profile.FullName = raw.Name
	profile.AvatarURL = raw.Picture

	return profile, nil
}
