package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/hash"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
	"golang.org/x/oauth2"
)

const DefaultRedirectPath = "/oauth/callback"

type StartInput struct {
	Provider     string `validate:"required"`
	RedirectPath string
}

type StartOutput struct {
	AuthURL string
}

type CallbackInput struct {
	Provider string `validate:"required"`
	Code     string `validate:"required"`
	State    string `validate:"required"`
}

type CallbackOutput struct {
	Provider     string
	RedirectPath string
	Profile      Profile
}

type ProviderConfig struct {
	Name            string
	ClientID        string
	ClientSecret    string
	AuthURL         string
	TokenURL        string
	UserInfoURL     string
	UserEmailURL    string
	Scopes          []string
	RequireEmail    bool
	RequireVerified bool
}

type Profile struct {
	ProviderUserID string
	Email          string
	EmailVerified  bool
	FullName       string
	AvatarURL      string
	Nickname       string
}

type Provider interface {
	Name() string
	ApplyDefaults(cfg ProviderConfig) ProviderConfig
	FetchProfile(ctx context.Context, client *http.Client, cfg ProviderConfig, accessToken string) (Profile, error)
}

type Repo interface {
	CreateOAuthState(ctx context.Context, in entity.OAuthState) error
	GetOAuthState(ctx context.Context, state string) (*entity.OAuthState, error)
	DeleteOAuthState(ctx context.Context, id int64) error
}

type Dependency struct {
	Repo       Repo
	Validator  validator.Validator
	Config     config.Config
	HMAC       hash.Hash
	UID        uid.NumberID
	OID        uid.StringID
	Clock      clock.Clocker
	HTTPClient *http.Client
	Providers  []Provider
}

type Service struct {
	repo       Repo
	validator  validator.Validator
	cfg        config.Config
	hmac       hash.Hash
	uid        uid.NumberID
	oid        uid.StringID
	clock      clock.Clocker
	httpClient *http.Client
	providers  map[string]Provider
}

func NewService(dep Dependency) *Service {
	client := dep.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &Service{
		repo:       dep.Repo,
		validator:  dep.Validator,
		cfg:        dep.Config,
		hmac:       dep.HMAC,
		uid:        dep.UID,
		oid:        dep.OID,
		clock:      dep.Clock,
		httpClient: client,
		providers:  buildProviders(dep.Providers),
	}
}

func (s *Service) Start(ctx context.Context, in StartInput) (*StartOutput, error) {
	in.Provider = strings.ToLower(strings.TrimSpace(in.Provider))
	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	_, providerCfg, err := s.providerConfig(in.Provider)
	if err != nil {
		return nil, err
	}

	redirectPath := sanitizeRedirectPath(in.RedirectPath)

	state := s.oid.Generate()
	verifier, err := GenerateRandomToken(32)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate pkce verifier", "error", err)
		return nil, goerror.NewServer(err)
	}

	stateHash, err := s.hmac.Hash(state)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash oauth state", "error", err)
		return nil, goerror.NewServer(err)
	}

	stateTTL := s.cfg.GetMinute("modules.identity.oauth.state_ttl_minutes")
	if stateTTL <= 0 {
		stateTTL = 10 * time.Minute
	}

	if err := s.repo.CreateOAuthState(ctx, entity.OAuthState{
		ID:           s.uid.Generate(),
		State:        string(stateHash),
		Provider:     in.Provider,
		CodeVerifier: verifier,
		RedirectPath: redirectPath,
		ExpiresAt:    s.clock.Now().Add(stateTTL),
	}); err != nil {
		slog.ErrorContext(ctx, "failed to repo create oauth state", "provider", in.Provider, "error", err)
		return nil, goerror.NewServer(err)
	}

	callbackURL, err := s.callbackURL(in.Provider)
	if err != nil {
		return nil, err
	}

	config := oauth2Config(providerCfg, callbackURL)
	url := config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	return &StartOutput{AuthURL: url}, nil
}

func (s *Service) Callback(ctx context.Context, in CallbackInput) (*CallbackOutput, error) {
	in.Provider = strings.ToLower(strings.TrimSpace(in.Provider))
	in.Code = strings.TrimSpace(in.Code)
	in.State = strings.TrimSpace(in.State)

	if err := s.validator.Validate(in); err != nil {
		return &CallbackOutput{RedirectPath: DefaultRedirectPath}, goerror.NewInvalidInput(err)
	}

	provider, providerCfg, err := s.providerConfig(in.Provider)
	if err != nil {
		return &CallbackOutput{RedirectPath: DefaultRedirectPath}, err
	}

	stateHash, err := s.hmac.Hash(in.State)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash oauth state", "error", err)
		return &CallbackOutput{RedirectPath: DefaultRedirectPath}, goerror.NewServer(err)
	}

	state, err := s.repo.GetOAuthState(ctx, string(stateHash))
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "oauth state not found or expired")
		return &CallbackOutput{RedirectPath: DefaultRedirectPath}, goerror.NewBusiness("invalid oauth state", goerror.CodeForbidden)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get oauth state", "error", err)
		return &CallbackOutput{RedirectPath: DefaultRedirectPath}, goerror.NewServer(err)
	}
	defer func() {
		if err := s.repo.DeleteOAuthState(ctx, state.ID); err != nil {
			slog.ErrorContext(ctx, "failed to delete oauth state", "state_id", state.ID, "error", err)
		}
	}()

	redirectPath := sanitizeRedirectPath(state.RedirectPath)

	if state.Provider != in.Provider {
		slog.WarnContext(ctx, "oauth provider mismatch", "expected", state.Provider, "got", in.Provider)
		return &CallbackOutput{RedirectPath: redirectPath}, goerror.NewBusiness("invalid oauth session", goerror.CodeForbidden)
	}

	callbackURL, err := s.callbackURL(in.Provider)
	if err != nil {
		return &CallbackOutput{RedirectPath: redirectPath}, err
	}

	accessToken, err := s.exchangeCode(ctx, providerCfg, in.Code, callbackURL, state.CodeVerifier)
	if err != nil {
		return &CallbackOutput{RedirectPath: redirectPath}, err
	}

	profile, err := s.fetchProfile(ctx, provider, providerCfg, accessToken)
	if err != nil {
		return &CallbackOutput{RedirectPath: redirectPath}, err
	}

	profile.Email = strings.TrimSpace(strings.ToLower(profile.Email))
	if providerCfg.RequireEmail && profile.Email == "" {
		return &CallbackOutput{RedirectPath: redirectPath}, goerror.NewBusiness("email not available from provider", goerror.CodeInvalidInput)
	}
	if providerCfg.RequireVerified && !profile.EmailVerified {
		return &CallbackOutput{RedirectPath: redirectPath}, goerror.NewBusiness("email not verified by provider", goerror.CodeForbidden)
	}

	return &CallbackOutput{
		Provider:     in.Provider,
		RedirectPath: redirectPath,
		Profile:      profile,
	}, nil
}

func (s *Service) providerConfig(provider string) (Provider, ProviderConfig, error) {
	if !s.cfg.GetBool("modules.identity.oauth.enabled") {
		return nil, ProviderConfig{}, goerror.NewBusiness("oauth login disabled", goerror.CodeForbidden)
	}

	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return nil, ProviderConfig{}, goerror.NewInvalidFormat("provider is required")
	}

	prov, ok := s.providers[provider]
	if !ok {
		return nil, ProviderConfig{}, goerror.NewInvalidFormat("provider not supported")
	}

	var cfg ProviderConfig
	base := "modules.identity.oauth.providers." + provider
	cfg.Name = provider
	cfg.ClientID = strings.TrimSpace(s.cfg.GetString(base + ".client_id"))
	cfg.ClientSecret = strings.TrimSpace(s.cfg.GetString(base + ".client_secret"))
	cfg.AuthURL = strings.TrimSpace(s.cfg.GetString(base + ".auth_url"))
	cfg.TokenURL = strings.TrimSpace(s.cfg.GetString(base + ".token_url"))
	cfg.UserInfoURL = strings.TrimSpace(s.cfg.GetString(base + ".userinfo_url"))
	cfg.UserEmailURL = strings.TrimSpace(s.cfg.GetString(base + ".user_email_url"))
	for _, scope := range s.cfg.GetArray(base + ".scopes") {
		scope = strings.TrimSpace(scope)
		if scope != "" {
			cfg.Scopes = append(cfg.Scopes, scope)
		}
	}

	cfg = prov.ApplyDefaults(cfg)

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, ProviderConfig{}, goerror.NewBusiness("oauth provider not configured", goerror.CodeForbidden)
	}

	return prov, cfg, nil
}

func (s *Service) callbackURL(provider string) (string, error) {
	base := strings.TrimSpace(s.cfg.GetString("app.api_base_url"))
	if base == "" {
		return "", goerror.NewBusiness("api base url not configured", goerror.CodeForbidden)
	}

	return strings.TrimRight(base, "/") + "/api/v1/identity/oauth/" + provider + "/callback", nil
}

func (s *Service) exchangeCode(ctx context.Context, provider ProviderConfig, code, redirectURI, verifier string) (string, error) {
	config := oauth2Config(provider, redirectURI)
	ctx = context.WithValue(ctx, oauth2.HTTPClient, s.httpClient)
	resp, err := config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		slog.WarnContext(ctx, "oauth token exchange failed", "error", err)
		return "", goerror.NewBusiness("oauth token exchange failed", goerror.CodeUnauthorized)
	}

	if resp == nil || resp.AccessToken == "" {
		return "", goerror.NewBusiness("oauth access token missing", goerror.CodeUnauthorized)
	}

	return resp.AccessToken, nil
}

func (s *Service) fetchProfile(ctx context.Context, provider Provider, cfg ProviderConfig, accessToken string) (Profile, error) {
	return provider.FetchProfile(ctx, s.httpClient, cfg, accessToken)
}

func pkceChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func GenerateRandomToken(size int) (string, error) {
	if size <= 0 {
		size = 32
	}
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func sanitizeRedirectPath(raw string) string {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return DefaultRedirectPath
	}
	if strings.Contains(clean, "://") {
		return DefaultRedirectPath
	}
	if !strings.HasPrefix(clean, "/") {
		return DefaultRedirectPath
	}
	return clean
}

func oauth2Config(provider ProviderConfig, redirectURI string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		RedirectURL:  redirectURI,
		Scopes:       provider.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  provider.AuthURL,
			TokenURL: provider.TokenURL,
		},
	}
}
