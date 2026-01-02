package usecase

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/shared/constant"
)

type UserUpdateInput struct {
	ID       int64             `validate:"required,gt=0"`
	Email    string            `validate:"omitempty,email"`
	Password string            `validate:"omitempty,password"`
	FullName string            `validate:"omitempty,min=5,max=100,alphaspace"`
	Status   entity.UserStatus `validate:"omitempty,gt=0"`
}

func (s *Usecase) UserUpdate(ctx context.Context, in UserUpdateInput) error {
	ctx, span := s.startSpan(ctx, "UserUpdate")
	defer span.End()

	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.FullName = strings.TrimSpace(in.FullName)

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm, err := s.authenticatedAndAuthorized(ctx, constant.PermIdentityMgmtUsers, constant.PermActUpdate)
	if err != nil {
		return err
	}

	user, err := s.repoDB.GetUserByID(ctx, in.ID, false)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user not found", "user_id", in.ID)
		return goerror.NewBusiness("user not found", goerror.CodeNotFound)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", in.ID, "error", err)
		return goerror.NewServer(err)
	}

	checkEmail, err := s.repoDB.GetUserByEmail(ctx, in.Email, true)
	if err == nil && checkEmail != nil && user.Email != checkEmail.Email {
		slog.WarnContext(ctx, "user account is already exists", "email", in.Email)
		return goerror.NewBusiness("user account with that email already exists", goerror.CodeConflict)
	}
	if !errors.Is(err, goerror.ErrNotFound) {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", in.Email, "error", err)
		return goerror.NewServer(err)
	}

	var newHash string
	if in.Password != "" {
		hash, err := s.bcrypt.Hash(in.Password)
		if err != nil {
			slog.ErrorContext(ctx, "failed to hash new password", "user_id", user.ID, "error", err)
			return goerror.NewServer(err)
		}
		newHash = string(hash)
	}

	patchUser := entity.PatchUser{
		ID:        user.ID,
		UpdatedBy: clm.UserID,
		Email:     in.Email,
		FullName:  in.FullName,
		Status:    in.Status.Ensure(),
	}
	if in.FullName != "" {
		patchUser.AvatarURL = "https://ui-avatars.com/api/?name=" + url.QueryEscape(in.FullName)
	}
	if err := s.repoDB.PatchUser(ctx, patchUser, newHash); err != nil {
		slog.ErrorContext(ctx, "failed to repo patch user", "user_id", user.ID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
