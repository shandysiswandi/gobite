package inbound

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/identity/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

// HTTPEndpoint exposes HTTP handlers for authentication and profile workflows.
type HTTPEndpoint struct {
	uc uc
}

// Login authenticates a user and returns tokens or an MFA challenge.
// @Summary Authenticate user
// @Description Validates credentials and returns access/refresh tokens. If MFA is required, a challenge ID is returned.
// @Tags Identity
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login payload"
// @Success 200 {object} router.successResponse{data=LoginResponse} "Authentication result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Invalid credentials"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/login [post]
func (h *HTTPEndpoint) Login(r *router.Request) (any, error) {
	var req LoginRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	resp, err := h.uc.Login(r.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return LoginResponse{
		AccessToken:      resp.AccessToken,
		RefreshToken:     resp.RefreshToken,
		MfaRequired:      resp.MfaRequired,
		ChallengeToken:   resp.ChallengeToken,
		AvailableMethods: resp.AvailableMethods,
	}, nil
}

// Login2FA completes an 2FA login challenge and issues tokens.
// @Summary Complete 2FA login
// @Description Verifies the 2FA code for a login challenge and returns access/refresh tokens.
// @Tags Identity
// @Accept json
// @Produce json
// @Param request body Login2FARequest true "2FA login payload"
// @Success 200 {object} router.successResponse{data=Login2FAResponse} "Authentication result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Invalid MFA code"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/login/2fa [post]
func (h *HTTPEndpoint) Login2FA(r *router.Request) (any, error) {
	var req Login2FARequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	resp, err := h.uc.Login2FA(r.Context(), usecase.Login2FAInput{
		ChallengeToken: req.ChallengeToken,
		Method:         entity.MFATypeFromString(req.Method),
		Code:           req.Code,
	})
	if err != nil {
		return nil, err
	}

	return Login2FAResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// RefreshToken issues a new access token using a refresh token.
// @Summary Refresh access token
// @Description Exchanges a refresh token for a new access/refresh token pair.
// @Tags Identity
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token payload"
// @Success 200 {object} router.successResponse{data=RefreshTokenResponse} "Token refresh result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Invalid refresh token"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/refresh [post]
func (h *HTTPEndpoint) RefreshToken(r *router.Request) (any, error) {
	var req RefreshTokenRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	resp, err := h.uc.RefreshToken(r.Context(), usecase.RefreshTokenInput{RefreshToken: req.RefreshToken})
	if err != nil {
		return nil, err
	}

	return RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// Register creates a new user account.
// @Summary Register user
// @Description Creates a new account and sends a verification email.
// @Tags Identity
// @Accept json
// @Param request body RegisterRequest true "Registration payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 409 {object} router.errorResponse "Email already registered"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/register [post]
func (h *HTTPEndpoint) Register(r *router.Request) (any, error) {
	var req RegisterRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	if err := h.uc.Register(r.Context(), usecase.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	}); err != nil {
		return nil, err
	}

	return &RegisterResponse{}, nil
}

// RegisterResend resends the account verification email if applicable.
// @Summary Resend verification email
// @Description Sends a new verification email when an account exists for the provided address.
// @Tags Identity
// @Accept json
// @Produce json
// @Param request body RegisterResendRequest true "Resend verification payload"
// @Success 200 {object} router.successResponse{data=RegisterResendResponse} "Resend result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/register/resend [post]
func (h *HTTPEndpoint) RegisterResend(r *router.Request) (any, error) {
	var req RegisterResendRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	if err := h.uc.RegisterResend(r.Context(), usecase.RegisterResendInput{
		Email: req.Email,
	}); err != nil {
		return nil, err
	}

	return &RegisterResendResponse{}, nil
}

// RegisterVerify verifies a user's email using a verification token.
// @Summary Verify email
// @Description Confirms the user's email address using the provided verification token.
// @Tags Identity
// @Accept json
// @Param request body EmailVerifyRequest true "Email verification payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 404 {object} router.errorResponse "Verification token not found"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/register/verify [post]
func (h *HTTPEndpoint) RegisterVerify(r *router.Request) (any, error) {
	var req EmailVerifyRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	return nil, h.uc.RegisterVerify(r.Context(), usecase.RegisterVerifyInput{ChallengeToken: req.ChallengeToken})
}

// PasswordForgot initiates a password reset flow.
// @Summary Request password reset
// @Description Sends password reset instructions to the provided email address.
// @Tags Identity
// @Accept json
// @Param request body PasswordForgotRequest true "Forgot password payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/password/forgot [post]
func (h *HTTPEndpoint) PasswordForgot(r *router.Request) (any, error) {
	var req PasswordForgotRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	if err := h.uc.PasswordForgot(r.Context(), usecase.PasswordForgotInput{Email: req.Email}); err != nil {
		return nil, err
	}

	return &PasswordForgotResponse{}, nil
}

// PasswordReset completes a password reset using a reset token.
// @Summary Reset password
// @Description Sets a new password using the provided reset token.
// @Tags Identity
// @Accept json
// @Param request body PasswordResetRequest true "Reset password payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 404 {object} router.errorResponse "Reset token not found"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/password/reset [post]
func (h *HTTPEndpoint) PasswordReset(r *router.Request) (any, error) {
	var req PasswordResetRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	return nil, h.uc.PasswordReset(r.Context(), usecase.PasswordResetInput{
		ChallengeToken: req.ChallengeToken,
		NewPassword:    req.NewPassword,
	})
}

// PasswordChange updates the current user's password.
// @Summary Change password
// @Description Updates the user's password after validating the current password.
// @Tags Identity
// @Accept json
// @Param request body PasswordChangeRequest true "Change password payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/password/change [post]
func (h *HTTPEndpoint) PasswordChange(r *router.Request) (any, error) {
	var req PasswordChangeRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	return nil, h.uc.PasswordChange(r.Context(), usecase.PasswordChangeInput{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
}

// Logout revokes a refresh token.
// @Summary Logout
// @Description Invalidates the provided refresh token.
// @Tags Identity
// @Accept json
// @Param request body LogoutRequest true "Logout payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/logout [post]
func (h *HTTPEndpoint) Logout(r *router.Request) (any, error) {
	var req LogoutRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	return nil, h.uc.Logout(r.Context(), usecase.LogoutInput{RefreshToken: req.RefreshToken})
}

// LogoutAll revokes all active sessions for the current user.
// @Summary Logout all sessions
// @Description Invalidates all refresh tokens for the authenticated user.
// @Tags Identity
// @Success 204 "No Content"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/logout-all [post]
func (h *HTTPEndpoint) LogoutAll(r *router.Request) (any, error) {
	return nil, h.uc.LogoutAll(r.Context(), usecase.LogoutAllInput{})
}

// TOTPSetup registers a new TOTP factor for the current user.
// @Summary Setup TOTP
// @Description Creates a TOTP factor and returns the shared secret and otpauth URI.
// @Tags Identity
// @Accept json
// @Produce json
// @Param request body TOTPSetupRequest true "TOTP setup payload"
// @Success 200 {object} router.successResponse{data=TOTPSetupResponse} "TOTP setup result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/mfa/totp/setup [post]
func (h *HTTPEndpoint) TOTPSetup(r *router.Request) (any, error) {
	var req TOTPSetupRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	resp, err := h.uc.TOTPSetup(r.Context(), usecase.TOTPSetupInput{
		FriendlyName:    req.FriendlyName,
		CurrentPassword: req.CurrentPassword,
	})
	if err != nil {
		return nil, err
	}

	return TOTPSetupResponse{
		ChallengeToken: resp.ChallengeToken,
		Key:            resp.Key,
		URI:            resp.URI,
	}, nil
}

// TOTPConfirm verifies a TOTP code to activate the factor.
// @Summary Confirm TOTP
// @Description Verifies the TOTP code and activates the MFA factor.
// @Tags Identity
// @Accept json
// @Param request body TOTPConfirmRequest true "TOTP confirmation payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 404 {object} router.errorResponse "MFA factor not found"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/mfa/totp/confirm [post]
func (h *HTTPEndpoint) TOTPConfirm(r *router.Request) (any, error) {
	var req TOTPConfirmRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	resp, err := h.uc.TOTPConfirm(r.Context(), usecase.TOTPConfirmInput{
		ChallengeToken: req.ChallengeToken,
		Code:           req.Code,
	})
	if err != nil {
		return nil, err
	}

	return &TOTPConfirmResponse{RecoveryCodes: resp.RecoveryCodes}, nil
}

// BackupCodeRotate rotates backup codes for the current user.
// @Summary Rotate backup codes
// @Description Generates a new set of recovery codes for the authenticated user.
// @Tags Identity
// @Accept json
// @Produce json
// @Param request body BackupCodeRotateRequest true "Backup code rotation payload"
// @Success 200 {object} router.successResponse{data=BackupCodeRotateResponse} "Backup codes rotated"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/mfa/backup_code/rotate [post]
func (h *HTTPEndpoint) BackupCodeRotate(r *router.Request) (any, error) {
	var req BackupCodeRotateRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	resp, err := h.uc.BackupCodeRotate(r.Context(), usecase.BackupCodeRotateInput{CurrentPassword: req.CurrentPassword})
	if err != nil {
		return nil, err
	}

	return &BackupCodeRotateResponse{RecoveryCodes: resp.RecoveryCodes}, nil
}

// ProfileUpdate updates the current user's profile information.
// @Summary Update profile
// @Description Updates profile details for the authenticated user.
// @Tags Identity
// @Accept json
// @Param request body UpdateProfileRequest true "Profile update payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/profile [put]
func (h *HTTPEndpoint) ProfileUpdate(r *router.Request) (any, error) {
	var req UpdateProfileRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	return nil, h.uc.ProfileUpdate(r.Context(), usecase.ProfileUpdateInput{FullName: req.FullName})
}

// ProfileUpdateAvatar updates the current user's avatar URL.
// @Summary Update profile avatar
// @Description Updates avatar for the authenticated user.
// @Tags Identity
// @Accept multipart/form-data
// @Param avatar formData file true "Avatar image"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/profile/avatar [put]
func (h *HTTPEndpoint) ProfileUpdateAvatar(r *router.Request) (any, error) {
	ctx := r.Context()

	file, err := r.StreamSingleFile("avatar")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close file", "error", err)
		}
	}()

	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, goerror.NewInvalidFormat()
	}

	return nil, h.uc.ProfileUpdateAvatar(ctx, usecase.ProfileUpdateAvatarInput{
		File:        io.MultiReader(bytes.NewReader(head[:n]), file),
		ContentType: http.DetectContentType(head[:n]),
	})
}

// Profile retrieves the current user's profile details.
// @Summary Get profile
// @Description Returns profile information for the authenticated user.
// @Tags Identity
// @Produce json
// @Success 200 {object} router.successResponse{data=ProfileResponse} "Profile result"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/profile [get]
func (h *HTTPEndpoint) Profile(r *router.Request) (any, error) {
	resp, err := h.uc.Profile(r.Context(), usecase.ProfileInput{})
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

// UserList returns a list of users with optional filters.
// @Summary List users
// @Description Returns a paginated list of users with optional search and status filters.
// @Tags Identity
// @Produce json
// @Param search query string false "Search by email or full name"
// @Param sort_by query string false "Sort by email, full name and etc."
// @Param sort_order query string false "Sort order asc or desc"
// @Param status query int false "Filter by status (1=unverified|2=active|3=banned|4=deleted)"
// @Param size query int false "Pagination size"
// @Param page query int false "Pagination page"
// @Success 200 {object} router.successResponse{data=UsersResponse} "User list"
// @Failure 400 {object} router.errorResponse "Invalid query parameters"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 403 {object} router.errorResponse "Forbidden"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/users [get]
func (h *HTTPEndpoint) UserList(r *router.Request) (any, error) {
	size, err := r.GetQueryInt32("size")
	if err != nil {
		return nil, err
	}

	page, err := r.GetQueryInt32("page")
	if err != nil {
		return nil, err
	}

	status, err := r.GetQueryInt16("status")
	if err != nil {
		return nil, err
	}

	resp, err := h.uc.UserList(r.Context(), usecase.UserListInput{
		Search:    r.GetQuery("search"),
		Status:    entity.UserStatus(status).Ensure(),
		SortBy:    r.GetQuery("sort_by"),
		SortOrder: r.GetQuery("sort_order"),
		Size:      size,
		Page:      page,
	})
	if err != nil {
		return nil, err
	}

	users := make([]UserResponse, 0, len(resp.Users))
	for _, item := range resp.Users {
		users = append(users, UserResponse{
			ID:        item.ID,
			Email:     item.Email,
			FullName:  item.FullName,
			AvatarURL: item.AvatarURL,
			Status:    item.Status,
			UpdateAt:  item.UpdatedAt,
		})
	}

	return UsersResponse{
		total: resp.Total,
		size:  resp.Size,
		page:  resp.Page,
		Users: users,
	}, nil
}
