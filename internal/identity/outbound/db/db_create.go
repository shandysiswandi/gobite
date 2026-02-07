package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

func (s *DB) CreateChallenge(ctx context.Context, in entity.Challenge) (err error) {
	ctx, span := s.startSpan(ctx, "CreateChallenge")
	defer func() { s.endSpan(span, err) }()

	err = s.mapError(s.query.CreateIdentityChallenge(ctx, sqlc.CreateIdentityChallengeParams{
		ID:        in.ID,
		UserID:    in.UserID,
		Token:     in.Token,
		Purpose:   in.Purpose,
		ExpiresAt: pgtype.Timestamptz{Valid: true, Time: in.ExpiresAt},
		Metadata:  in.Metadata,
	}))
	return err
}

func (s *DB) CreateRefreshToken(ctx context.Context, in entity.RefreshToken) (err error) {
	ctx, span := s.startSpan(ctx, "CreateRefreshToken")
	defer func() { s.endSpan(span, err) }()

	err = s.mapError(s.query.CreateIdentityRefreshToken(ctx, sqlc.CreateIdentityRefreshTokenParams{
		ID:        in.ID,
		UserID:    in.UserID,
		Token:     in.Token,
		ExpiresAt: pgtype.Timestamptz{Valid: true, Time: in.ExpiresAt},
		Metadata:  in.Metadata,
	}))
	return err
}

func (s *DB) CreateUserConnection(ctx context.Context, in entity.UserConnection) (err error) {
	ctx, span := s.startSpan(ctx, "CreateUserConnection")
	defer func() { s.endSpan(span, err) }()

	err = s.mapError(s.query.CreateIdentityUserConnection(ctx, sqlc.CreateIdentityUserConnectionParams{
		ID:             in.ID,
		UserID:         in.UserID,
		Provider:       in.Provider,
		ProviderUserID: in.ProviderUserID,
	}))
	return err
}

func (s *DB) CreateOAuthState(ctx context.Context, in entity.OAuthState) (err error) {
	ctx, span := s.startSpan(ctx, "CreateOAuthState")
	defer func() { s.endSpan(span, err) }()

	err = s.mapError(s.query.CreateIdentityOAuthState(ctx, sqlc.CreateIdentityOAuthStateParams{
		ID:           in.ID,
		State:        in.State,
		Provider:     in.Provider,
		CodeVerifier: in.CodeVerifier,
		RedirectPath: in.RedirectPath,
		ExpiresAt:    pgtype.Timestamptz{Valid: true, Time: in.ExpiresAt},
	}))
	return err
}
