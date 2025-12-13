package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) Profile(ctx context.Context, in domain.ProfileInput) (*domain.ProfileOutput, error) {
	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	user, err := s.getUserByEmail(ctx, clm.Payload().Email)
	if err != nil {
		return nil, err
	}

	return &domain.ProfileOutput{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
		Status:    user.Status.String(),
	}, nil
}
