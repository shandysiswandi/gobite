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

type (
	UserCreateInput struct {
		Email    string            `validate:"required,email"`
		Password string            `validate:"required,password"`
		FullName string            `validate:"required,min=5,max=100,alphaspace"`
		Status   entity.UserStatus `validate:"required,gt=0"`
	}
)

func (s *Usecase) UserCreate(ctx context.Context, in UserCreateInput) error {
	ctx, span := s.startSpan(ctx, "UserCreate")
	defer span.End()

	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.FullName = strings.TrimSpace(in.FullName)

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm, err := s.authenticatedAndAuthorized(ctx, constant.PermIdentityMgmtUsers, constant.PermActCreate)
	if err != nil {
		return err
	}

	user, err := s.repoDB.GetUserByEmail(ctx, in.Email, true)
	if err == nil && user != nil {
		slog.WarnContext(ctx, "user account is already exists", "email", in.Email)
		return goerror.NewBusiness("user account with that email already exists", goerror.CodeConflict)
	}
	if !errors.Is(err, goerror.ErrNotFound) {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", in.Email, "error", err)
		return goerror.NewServer(err)
	}

	hashedPassword, err := s.bcrypt.Hash(in.Password)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password", "error", err)
		return goerror.NewServer(err)
	}

	newUser := entity.NewUser{
		ID:        s.uid.Generate(),
		Email:     in.Email,
		FullName:  in.FullName,
		AvatarURL: "https://ui-avatars.com/api/?name=" + url.QueryEscape(in.FullName),
		Status:    in.Status,
		CreatedBy: clm.UserID,
		UpdatedBy: clm.UserID,
	}

	if err := s.repoDB.NewUser(ctx, newUser, string(hashedPassword)); err != nil {
		slog.ErrorContext(ctx, "failed to repo create new user", "new_user", newUser, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
