package inbound

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/shandysiswandi/gobite/internal/auth/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

// HTTPEndpoint exposes HTTP handlers for authentication and profile workflows.
type HTTPEndpoint struct {
	uc uc
}

// Login authenticates a user and returns tokens or an MFA challenge.
// @Summary Authenticate user
// @Description Validates credentials and returns access/refresh tokens. If MFA is required, a challenge ID is returned.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login payload"
// @Success 200 {object} router.successResponse{data=LoginResponse} "Authentication result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Invalid credentials"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *HTTPEndpoint) Login(ctx context.Context, r *http.Request) (any, error) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	resp, err := h.uc.Login(ctx, usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return LoginResponse{
		AccessToken:    resp.AccessToken,
		RefreshToken:   resp.RefreshToken,
		MfaRequired:    resp.MfaRequired,
		ChallengeToken: resp.ChallengeToken,
	}, nil
}

// LoginMFA completes an MFA login challenge and issues tokens.
// @Summary Complete MFA login
// @Description Verifies the MFA code for a login challenge and returns access/refresh tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginMFARequest true "MFA login payload"
// @Success 200 {object} router.successResponse{data=LoginMFAResponse} "Authentication result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Invalid MFA code"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/login/mfa [post]
func (h *HTTPEndpoint) LoginMFA(ctx context.Context, r *http.Request) (any, error) {
	var req LoginMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	resp, err := h.uc.LoginMFA(ctx, usecase.LoginMFAInput{
		ChallengeToken: req.ChallengeToken,
		Code:           req.Code,
	})
	if err != nil {
		return nil, err
	}

	return LoginMFAResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// Register creates a new user account.
// @Summary Register user
// @Description Creates a new account and sends a verification email.
// @Tags Auth
// @Accept json
// @Param request body RegisterRequest true "Registration payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 409 {object} router.errorResponse "Email already registered"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *HTTPEndpoint) Register(ctx context.Context, r *http.Request) (any, error) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.Register(ctx, usecase.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// RegisterResend resends the account verification email if applicable.
// @Summary Resend verification email
// @Description Sends a new verification email when an account exists for the provided address.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterResendRequest true "Resend verification payload"
// @Success 200 {object} router.successResponse{data=RegisterResendesponse} "Resend result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/register/resend [post]
func (h *HTTPEndpoint) RegisterResend(ctx context.Context, r *http.Request) (any, error) {
	var req RegisterResendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.RegisterResend(ctx, usecase.RegisterResendInput{
		Email: req.Email,
		IP:    r.RemoteAddr,
	}); err != nil {
		return nil, err
	}

	return &RegisterResendesponse{}, nil
}

// EmailVerify verifies a user's email using a verification token.
// @Summary Verify email
// @Description Confirms the user's email address using the provided verification token.
// @Tags Auth
// @Accept json
// @Param request body EmailVerifyRequest true "Email verification payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 404 {object} router.errorResponse "Verification token not found"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/email/verify [post]
func (h *HTTPEndpoint) EmailVerify(ctx context.Context, r *http.Request) (any, error) {
	var req EmailVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.EmailVerify(ctx, usecase.EmailVerifyInput{Token: req.Token}); err != nil {
		return nil, err
	}

	return nil, nil
}

// ForgotPassword initiates a password reset flow.
// @Summary Request password reset
// @Description Sends password reset instructions to the provided email address.
// @Tags Auth
// @Accept json
// @Param request body ForgotPasswordRequest true "Forgot password payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/password/forgot [post]
func (h *HTTPEndpoint) ForgotPassword(ctx context.Context, r *http.Request) (any, error) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.ForgotPassword(ctx, usecase.ForgotPasswordInput{Email: req.Email}); err != nil {
		return nil, err
	}

	return nil, nil
}

// ResetPassword completes a password reset using a reset token.
// @Summary Reset password
// @Description Sets a new password using the provided reset token.
// @Tags Auth
// @Accept json
// @Param request body ResetPasswordRequest true "Reset password payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 404 {object} router.errorResponse "Reset token not found"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/password/reset [post]
func (h *HTTPEndpoint) ResetPassword(ctx context.Context, r *http.Request) (any, error) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.ResetPassword(ctx, usecase.ResetPasswordInput{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// Logout revokes a refresh token.
// @Summary Logout
// @Description Invalidates the provided refresh token.
// @Tags Auth
// @Accept json
// @Param request body LogoutRequest true "Logout payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/logout [post]
func (h *HTTPEndpoint) Logout(ctx context.Context, r *http.Request) (any, error) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.Logout(ctx, usecase.LogoutInput{RefreshToken: req.RefreshToken}); err != nil {
		return nil, err
	}

	return nil, nil
}

// LogoutAll revokes all active sessions for the current user.
// @Summary Logout all sessions
// @Description Invalidates all refresh tokens for the authenticated user.
// @Tags Auth
// @Success 204 "No Content"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/logout-all [post]
func (h *HTTPEndpoint) LogoutAll(ctx context.Context, r *http.Request) (any, error) {
	if err := h.uc.LogoutAll(ctx, usecase.LogoutAllInput{}); err != nil {
		return nil, err
	}

	return nil, nil
}

// ChangePassword updates the current user's password.
// @Summary Change password
// @Description Updates the user's password after validating the current password.
// @Tags Auth
// @Accept json
// @Param request body ChangePasswordRequest true "Change password payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/password/change [post]
func (h *HTTPEndpoint) ChangePassword(ctx context.Context, r *http.Request) (any, error) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.ChangePassword(ctx, usecase.ChangePasswordInput{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// SetupTOTP registers a new TOTP factor for the current user.
// @Summary Setup TOTP
// @Description Creates a TOTP factor and returns the shared secret and otpauth URI.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SetupTOTPRequest true "TOTP setup payload"
// @Success 200 {object} router.successResponse{data=SetupTOTPResponse} "TOTP setup result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/mfa/totp/setup [post]
func (h *HTTPEndpoint) SetupTOTP(ctx context.Context, r *http.Request) (any, error) {
	var req SetupTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	resp, err := h.uc.SetupTOTP(ctx, usecase.SetupTOTPInput{
		FriendlyName:    req.FriendlyName,
		CurrentPassword: req.CurrentPassword,
	})
	if err != nil {
		return nil, err
	}

	return SetupTOTPResponse{
		ChallengeToken: resp.ChallengeToken,
		Key:            resp.Key,
		URI:            resp.URI,
	}, nil
}

// ConfirmTOTP verifies a TOTP code to activate the factor.
// @Summary Confirm TOTP
// @Description Verifies the TOTP code and activates the MFA factor.
// @Tags Auth
// @Accept json
// @Param request body ConfirmTOTPRequest true "TOTP confirmation payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 404 {object} router.errorResponse "MFA factor not found"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/mfa/totp/confirm [post]
func (h *HTTPEndpoint) ConfirmTOTP(ctx context.Context, r *http.Request) (any, error) {
	var req ConfirmTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.ConfirmTOTP(ctx, usecase.ConfirmTOTPInput{
		ChallengeToken: req.ChallengeToken,
		Code:           req.Code,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// RefreshToken issues a new access token using a refresh token.
// @Summary Refresh access token
// @Description Exchanges a refresh token for a new access/refresh token pair.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token payload"
// @Success 200 {object} router.successResponse{data=RefreshTokenResponse} "Token refresh result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Invalid refresh token"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
func (h *HTTPEndpoint) RefreshToken(ctx context.Context, r *http.Request) (any, error) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	resp, err := h.uc.RefreshToken(ctx, usecase.RefreshTokenInput{RefreshToken: req.RefreshToken})
	if err != nil {
		return nil, err
	}

	return RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// UpdateProfile updates the current user's profile information.
// @Summary Update profile
// @Description Updates profile details for the authenticated user.
// @Tags Auth
// @Accept json
// @Param request body UpdateProfileRequest true "Profile update payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/profile [put]
func (h *HTTPEndpoint) UpdateProfile(ctx context.Context, r *http.Request) (any, error) {
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.UpdateProfile(ctx, usecase.UpdateProfileInput{FullName: req.FullName}); err != nil {
		return nil, err
	}

	return nil, nil
}

// Profile retrieves the current user's profile details.
// @Summary Get profile
// @Description Returns profile information for the authenticated user.
// @Tags Auth
// @Produce json
// @Success 200 {object} router.successResponse{data=ProfileResponse} "Profile result"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/profile [get]
func (h *HTTPEndpoint) Profile(ctx context.Context, r *http.Request) (any, error) {
	resp, err := h.uc.Profile(ctx, usecase.ProfileInput{})
	if err != nil {
		return nil, err
	}

	return ProfileResponse{
		ID:        resp.ID,
		Email:     resp.Email,
		FullName:  resp.FullName,
		AvatarURL: resp.AvatarURL,
		Status:    resp.Status,
	}, nil
}
