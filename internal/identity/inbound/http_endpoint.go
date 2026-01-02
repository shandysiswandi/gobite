package inbound

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

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
// @Description Validates credentials and returns access/refresh tokens. If MFA is required, a challenge is returned.
// @Tags Identity, Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login payload"
// @Success 200 {object} router.successResponse{data=LoginResponse} "Authentication result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Invalid credentials"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error" example:{"message":"Login failed due to server error","error":{"detail":"Please try again later"}}
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
// @Tags Identity, Authentication
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
// @Tags Identity, Authentication
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
// @Tags Identity, Authentication
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
// @Tags Identity, Authentication
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
// @Tags Identity, Authentication
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
// @Tags Identity, Authentication
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
// @Tags Identity, Authentication
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
// @Tags Identity, Profile Security
// @Security BearerAuth
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
// @Tags Identity, Authentication
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
// @Tags Identity, Profile Security
// @Security BearerAuth
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
// @Tags Identity, Profile Security
// @Security BearerAuth
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
// @Tags Identity, Profile Security
// @Security BearerAuth
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

	if err := h.uc.TOTPConfirm(r.Context(), usecase.TOTPConfirmInput{
		ChallengeToken: req.ChallengeToken,
		Code:           req.Code,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// BackupCode rotates backup codes for the current user.
// @Summary Rotate backup codes
// @Description Generates a new set of recovery codes for the authenticated user.
// @Tags Identity, Profile Security
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body BackupCodeRequest true "Backup code rotation payload"
// @Success 200 {object} router.successResponse{data=BackupCodeResponse} "Backup codes rotated"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/mfa/backup_code/rotate [post]
func (h *HTTPEndpoint) BackupCode(r *router.Request) (any, error) {
	var req BackupCodeRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	resp, err := h.uc.BackupCode(r.Context(), usecase.BackupCodeInput{CurrentPassword: req.CurrentPassword})
	if err != nil {
		return nil, err
	}

	return &BackupCodeResponse{RecoveryCodes: resp.RecoveryCodes}, nil
}

// ProfileUpdate updates the current user's profile information.
// @Summary Update profile
// @Description Updates profile details for the authenticated user.
// @Tags Identity, Profile
// @Security BearerAuth
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
// @Tags Identity, Profile
// @Security BearerAuth
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
// @Tags Identity, Profile
// @Security BearerAuth
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

// @Summary Get profile permissions
// @Tags Identity, Profile Security
// @Security BearerAuth
// @Success 200 {object} router.successResponse{data=ProfilePermissionsResponse} "Permissions list"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/profile/permissions [get]
func (h *HTTPEndpoint) ProfilePermissions(r *router.Request) (any, error) {
	resp, err := h.uc.ProfilePermissions(r.Context())
	if err != nil {
		return nil, err
	}

	if resp == nil {
		resp = map[string][]string{}
	}

	return ProfilePermissionsResponse{Permissions: resp}, nil
}

// @Summary Get profile MFA settings
// @Description Returns MFA settings for the authenticated user.
// @Tags Identity, Profile Security
// @Security BearerAuth
// @Produce json
// @Success 200 {object} router.successResponse{data=ProfileSettingMFAResponse} "MFA settings"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/profile/settings/mfa [get]
func (h *HTTPEndpoint) ProfileSettingMFA(r *router.Request) (any, error) {
	resp, err := h.uc.ProfileSettingMFA(r.Context())
	if err != nil {
		return nil, err
	}

	return ProfileSettingMFAResponse{
		TOTPEnabled:       resp.TOTPEnabled,
		BackupCodeEnabled: resp.BackupCodeEnabled,
		SMSEnabled:        resp.SMSEnabled,
	}, nil
}

// UserList returns a list of users with optional filters.
// @Summary List users
// @Description Returns a paginated list of users with optional search and status filters.
// @Tags Identity, Management Users
// @Security BearerAuth
// @Produce json
// @Param search query string false "Search by email or full name"
// @Param sort_by query string false "Sort by email, full name and etc."
// @Param sort_order query string false "Sort order asc or desc"
// @Param status query []int false "Filter by statuses (1=unverified|2=active|3=banned|4=deleted)"
// @Param date_from query string false "Filter by created_at >= date_from (RFC3339)"
// @Param date_to query string false "Filter by created_at <= date_to (RFC3339)"
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

	dateFrom, err := r.GetQueryDate("date_from", time.RFC3339)
	if err != nil {
		return nil, err
	}

	dateTo, err := r.GetQueryDate("date_to", time.RFC3339)
	if err != nil {
		return nil, err
	}

	if !dateFrom.IsZero() && !dateTo.IsZero() && dateFrom.After(dateTo) {
		return nil, goerror.NewInvalidFormat("date_from must be before date_to")
	}

	resp, err := h.uc.UserList(r.Context(), usecase.UserListInput{
		Search:    r.GetQuery("search"),
		Statuses:  r.GetQueries("status"),
		SortBy:    r.GetQuery("sort_by"),
		SortOrder: r.GetQuery("sort_order"),
		DateFrom:  dateFrom,
		DateTo:    dateTo,
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

// @Summary Get user detail
// @Description Returns user details for a given user ID.
// @Tags Identity, Management Users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} router.successResponse{data=UserDetailResponse} "User detail"
// @Failure 400 {object} router.errorResponse "Invalid path parameter"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 403 {object} router.errorResponse "Forbidden"
// @Failure 404 {object} router.errorResponse "User not found"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/users/{id} [get]
func (h *HTTPEndpoint) UserDetail(r *router.Request) (any, error) {
	id, err := r.GetParamInt64("id")
	if err != nil {
		return nil, err
	}

	resp, err := h.uc.UserDetail(r.Context(), usecase.UserDetailInput{ID: id})
	if err != nil {
		return nil, err
	}

	return UserDetailResponse{User: UserResponse{
		ID:        resp.User.ID,
		Email:     resp.User.Email,
		FullName:  resp.User.FullName,
		AvatarURL: resp.User.AvatarURL,
		Status:    resp.User.Status,
		UpdateAt:  resp.User.UpdatedAt,
	}}, nil
}

// @Summary Create user
// @Description Creates a new user.
// @Tags Identity, Management Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UserCreateRequest true "User creation payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 403 {object} router.errorResponse "Forbidden"
// @Failure 409 {object} router.errorResponse "Email already registered"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/users [post]
func (h *HTTPEndpoint) UserCreate(r *router.Request) (any, error) {
	var req UserCreateRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	if err := h.uc.UserCreate(r.Context(), usecase.UserCreateInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Status:   req.Status,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// @Summary Update user
// @Description Updates a user by ID.
// @Tags Identity, Management Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body UserUpdateRequest true "User update payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 403 {object} router.errorResponse "Forbidden"
// @Failure 404 {object} router.errorResponse "User not found"
// @Failure 409 {object} router.errorResponse "Email already registered"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/users/{id} [put]
func (h *HTTPEndpoint) UserUpdate(r *router.Request) (any, error) {
	id, err := r.GetParamInt64("id")
	if err != nil {
		return nil, err
	}

	var req UserUpdateRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	if err := h.uc.UserUpdate(r.Context(), usecase.UserUpdateInput{
		ID:       id,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Status:   req.Status,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// @Summary Delete user
// @Description Deletes a user by ID.
// @Tags Identity, Management Users
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid path parameter"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 403 {object} router.errorResponse "Forbidden"
// @Failure 404 {object} router.errorResponse "User not found"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/users/{id} [delete]
func (h *HTTPEndpoint) UserDelete(r *router.Request) (any, error) {
	id, err := r.GetParamInt64("id")
	if err != nil {
		return nil, err
	}

	if err := h.uc.UserDelete(r.Context(), usecase.UserDeleteInput{ID: id}); err != nil {
		return nil, err
	}

	return nil, nil
}

// @Summary Export users
// @Description Returns user list for export with optional filters.
// @Tags Identity, Management Users
// @Security BearerAuth
// @Produce json
// @Param search query string false "Search by email or full name"
// @Param status query []int false "Filter by user status"
// @Param sort_by query string false "Sort by email, full name and etc."
// @Param sort_order query string false "Sort order: asc, desc"
// @Param date_from query string false "Filter by created_at >= date_from (RFC3339)"
// @Param date_to query string false "Filter by created_at <= date_to (RFC3339)"
// @Success 200 {object} router.successResponse{data=UsersResponse} "User export"
// @Failure 400 {object} router.errorResponse "Invalid query parameter"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 403 {object} router.errorResponse "Forbidden"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/users-export [get]
func (h *HTTPEndpoint) UserExport(r *router.Request) (any, error) {
	dateFrom, err := r.GetQueryDate("date_from", time.RFC3339)
	if err != nil {
		return nil, err
	}

	dateTo, err := r.GetQueryDate("date_to", time.RFC3339)
	if err != nil {
		return nil, err
	}

	if !dateFrom.IsZero() && !dateTo.IsZero() && dateFrom.After(dateTo) {
		return nil, goerror.NewInvalidFormat("date_from must be before date_to")
	}

	resp, err := h.uc.UserExport(r.Context(), usecase.UserExportInput{
		Search:    r.GetQuery("search"),
		Statuses:  r.GetQueries("status"),
		SortBy:    r.GetQuery("sort_by"),
		SortOrder: r.GetQuery("sort_order"),
		DateFrom:  dateFrom,
		DateTo:    dateTo,
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

	return UserExportResponse{Users: users}, nil
}

// @Summary Import users
// @Description Imports users in bulk.
// @Tags Identity, Management Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UserImportRequest true "User import payload"
// @Success 200 {object} router.successResponse{data=UserImportResponse} "User import result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 403 {object} router.errorResponse "Forbidden"
// @Failure 409 {object} router.errorResponse "Email already registered"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/identity/users-import [post]
func (h *HTTPEndpoint) UserImport(r *router.Request) (any, error) {
	var req UserImportRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	users := make([]usecase.UserImportUserInput, 0, len(req))
	for _, item := range req {
		users = append(users, usecase.UserImportUserInput{
			Email:    item.Email,
			Password: item.Password,
			FullName: item.FullName,
			Status:   item.Status,
		})
	}

	resp, err := h.uc.UserImport(r.Context(), usecase.UserImportInput{Users: users})
	if err != nil {
		return nil, err
	}

	return UserImportResponse{
		Created: resp.Created,
		Updated: resp.Updated,
	}, nil
}
