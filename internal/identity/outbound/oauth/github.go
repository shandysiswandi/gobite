package oauth

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type GitHubProvider struct{}

func (GitHubProvider) Name() string {
	return "github"
}

func (GitHubProvider) ApplyDefaults(cfg ProviderConfig) ProviderConfig {
	cfg.RequireEmail = true
	cfg.RequireVerified = true
	if cfg.AuthURL == "" {
		cfg.AuthURL = "https://github.com/login/oauth/authorize"
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = "https://github.com/login/oauth/access_token"
	}
	if cfg.UserInfoURL == "" {
		cfg.UserInfoURL = "https://api.github.com/user"
	}
	if cfg.UserEmailURL == "" {
		cfg.UserEmailURL = "https://api.github.com/user/emails"
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"read:user", "user:email"}
	}
	return cfg
}

func (GitHubProvider) FetchProfile(ctx context.Context, client *http.Client, cfg ProviderConfig, accessToken string) (Profile, error) {
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
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return profile, goerror.NewServer(err)
	}

	profile.ProviderUserID = strconv.FormatInt(raw.ID, 10)
	profile.Email = raw.Email
	profile.FullName = raw.Name
	profile.Nickname = raw.Login
	profile.AvatarURL = raw.AvatarURL

	if profile.Email == "" {
		email, verified, err := fetchGitHubPrimaryEmail(ctx, client, cfg, accessToken)
		if err != nil {
			return profile, err
		}
		profile.Email = email
		profile.EmailVerified = verified
	}

	return profile, nil
}

func fetchGitHubPrimaryEmail(ctx context.Context, client *http.Client, cfg ProviderConfig, accessToken string) (string, bool, error) {
	if cfg.UserEmailURL == "" {
		return "", false, goerror.NewBusiness("github email endpoint not configured", goerror.CodeInvalidFormat)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserEmailURL, nil)
	if err != nil {
		return "", false, goerror.NewServer(err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", false, goerror.NewServer(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, goerror.NewServer(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.WarnContext(ctx, "oauth github emails failed", "status", resp.StatusCode, "body", string(body))
		return "", false, goerror.NewBusiness("oauth user email lookup failed", goerror.CodeUnauthorized)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", false, goerror.NewServer(err)
	}

	for _, item := range emails {
		if item.Primary {
			return item.Email, item.Verified, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, emails[0].Verified, nil
	}

	return "", false, goerror.NewBusiness("email not available from provider", goerror.CodeInvalidInput)
}
