package inbound

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

type HTTPEndpoint struct {
	uc usecase
}

func (h *HTTPEndpoint) Login(ctx context.Context, r *http.Request) (any, error) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pkgerror.NewInvalidFormat()
	}

	resp, err := h.uc.Login(ctx, domain.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		MfaRequired:  resp.MfaRequired,
		PreAuthToken: resp.PreAuthToken,
	}, nil
}

func (h *HTTPEndpoint) Login2FA(ctx context.Context, r *http.Request) (any, error) {
	var req Login2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pkgerror.NewInvalidFormat()
	}

	resp, err := h.uc.Login2FA(ctx, domain.Login2FAInput{
		PreAuthToken: req.PreAuthToken,
		Code:         req.Code,
	})
	if err != nil {
		return nil, err
	}

	return Login2FAResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (h *HTTPEndpoint) Register(ctx context.Context, r *http.Request) (any, error) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pkgerror.NewInvalidFormat()
	}

	resp, err := h.uc.Register(ctx, domain.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		return nil, err
	}

	return RegisterResponse{IsNeedVerify: resp.IsNeedVerify}, nil
}

func (h *HTTPEndpoint) EmailVerify(ctx context.Context, r *http.Request) (any, error) {
	var req EmailVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pkgerror.NewInvalidFormat()
	}

	if err := h.uc.EmailVerify(ctx, domain.EmailVerifyInput{Token: req.Token}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *HTTPEndpoint) ForgotPassword(ctx context.Context, r *http.Request) (any, error) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pkgerror.NewInvalidFormat()
	}

	resp, err := h.uc.ForgotPassword(ctx, domain.ForgotPasswordInput{
		Email: req.Email,
	})
	if err != nil {
		return nil, err
	}

	return ForgotPasswordResponse{Success: resp.Success}, nil
}

func (h *HTTPEndpoint) ResetPassword(ctx context.Context, r *http.Request) (any, error) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pkgerror.NewInvalidFormat()
	}

	if err := h.uc.ResetPassword(ctx, domain.ResetPasswordInput{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *HTTPEndpoint) Logout(ctx context.Context, r *http.Request) (any, error) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pkgerror.NewInvalidFormat()
	}

	if err := h.uc.Logout(ctx, domain.LogoutInput{RefreshToken: req.RefreshToken}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *HTTPEndpoint) ChangePassword(ctx context.Context, r *http.Request) (any, error) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pkgerror.NewInvalidFormat()
	}

	if err := h.uc.ChangePassword(ctx, domain.ChangePasswordInput{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *HTTPEndpoint) RefreshToken(ctx context.Context, r *http.Request) (any, error) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pkgerror.NewInvalidFormat()
	}

	resp, err := h.uc.RefreshToken(ctx, domain.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}

	return RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (h *HTTPEndpoint) Profile(ctx context.Context, r *http.Request) (any, error) {
	resp, err := h.uc.Profile(ctx, domain.ProfileInput{})
	if err != nil {
		return nil, pkgerror.NewInvalidFormat()
	}

	return ProfileResponse{
		ID:        resp.ID,
		Email:     resp.Email,
		FullName:  resp.FullName,
		AvatarURL: resp.AvatarURL,
		Status:    resp.Status,
	}, nil
}
