package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

type ProfileInput struct{}

type ProfileOutput struct {
	ID        int64
	Email     string
	FullName  string
	AvatarURL string
	Status    string
}

func (s *Usecase) Profile(ctx context.Context, in ProfileInput) (*ProfileOutput, error) {
	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	user, err := s.getUserByID(ctx, clm.Payload().UserID)
	if err != nil {
		return nil, err
	}

	return &ProfileOutput{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
		Status:    user.Status.String(),
	}, nil
}
