package usecase

import (
	"context"
	"log/slog"
	"net/url"
	"strings"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/shared/constant"
)

type (
	UserImportUserInput struct {
		Email    string            `validate:"required,email"`
		Password string            `validate:"omitempty,password"`
		FullName string            `validate:"omitempty,min=5,max=100,alphaspace"`
		Status   entity.UserStatus `validate:"omitempty,gt=0"`
	}

	UserImportInput struct {
		Users []UserImportUserInput `validate:"required,min=1,max=10000,unique=Email,dive"`
	}

	UserImportOutput struct {
		Created int
		Updated int
	}
)

func (s *Usecase) UserImport(ctx context.Context, in UserImportInput) (*UserImportOutput, error) {
	ctx, span := s.startSpan(ctx, "UserImport")
	defer span.End()

	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	clm, err := s.authenticatedAndAuthorized(ctx, constant.PermIdentityMgmtUsers, constant.PermActCreate)
	if err != nil {
		return nil, err
	}

	users := make([]entity.UpsertUser, 0, len(in.Users))
	hashes := make(map[string]string, len(in.Users))
	for _, item := range in.Users {
		email := strings.TrimSpace(strings.ToLower(item.Email))
		fullName := strings.TrimSpace(item.FullName)

		if item.Password != "" {
			hash, err := s.bcrypt.Hash(item.Password)
			if err != nil {
				slog.ErrorContext(ctx, "failed to hash new password", "email", item.Email, "error", err)
				return nil, goerror.NewServer(err)
			}
			hashes[email] = string(hash)
		}

		upsertUser := entity.UpsertUser{
			ID:        s.uid.Generate(),
			CreatedBy: clm.UserID,
			UpdatedBy: clm.UserID,
			Email:     email,
			FullName:  fullName,
			Status:    item.Status,
		}
		if fullName != "" {
			upsertUser.AvatarURL = "https://ui-avatars.com/api/?name=" + url.QueryEscape(fullName)
		}

		users = append(users, upsertUser)
	}

	created, updated, err := s.repoDB.UpsertUsers(ctx, users, hashes)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo upsert users", "error", err)
		return nil, goerror.NewServer(err)
	}

	return &UserImportOutput{Created: created, Updated: updated}, nil
}
