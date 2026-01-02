package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

func (s *DB) RevokeRefreshToken(ctx context.Context, token string) (err error) {
	ctx, span := s.startSpan(ctx, "RevokeRefreshToken")
	defer func() { s.endSpan(span, err) }()

	return s.mapError(s.query.RevokeIdentityRefreshToken(ctx, token))
}

func (s *DB) RevokeAllRefreshToken(ctx context.Context, userID int64) (err error) {
	ctx, span := s.startSpan(ctx, "RevokeAllRefreshToken")
	defer func() { s.endSpan(span, err) }()

	return s.mapError(s.query.RevokeAllIdentityRefreshToken(ctx, userID))
}

func (s *DB) MarkMFABackupCodeUsed(ctx context.Context, bcID, userID int64) (_ bool, err error) {
	ctx, span := s.startSpan(ctx, "MarkMFABackupCodeUsed")
	defer func() { s.endSpan(span, err) }()

	rows, err := s.query.MarkIdentityMFABackupCodeUsed(ctx, sqlc.MarkIdentityMFABackupCodeUsedParams{
		UserID: userID,
		ID:     bcID,
	})
	if err != nil {
		return false, s.mapError(err)
	}

	return rows == 1, nil
}

func (s *DB) UpdateMFALastUsedAt(ctx context.Context, factorID, userID int64) (err error) {
	ctx, span := s.startSpan(ctx, "UpdateMFALastUsedAt")
	defer func() { s.endSpan(span, err) }()

	return s.mapError(s.query.UpdateIdentityMFALastUsedAt(ctx, sqlc.UpdateIdentityMFALastUsedAtParams{
		ID:     factorID,
		UserID: userID,
	}))
}

func (s *DB) UpdateUserProfile(ctx context.Context, id int64, fullName string) (err error) {
	ctx, span := s.startSpan(ctx, "UpdateUserProfile")
	defer func() { s.endSpan(span, err) }()

	return s.mapError(s.query.UpdateIdentityUserName(ctx, sqlc.UpdateIdentityUserNameParams{
		ID:        id,
		FullName:  fullName,
		UpdatedBy: id,
	}))
}

func (s *DB) UpdateUserAvatar(ctx context.Context, id int64, avatarURL string) (err error) {
	ctx, span := s.startSpan(ctx, "UpdateUserAvatar")
	defer func() { s.endSpan(span, err) }()

	return s.mapError(s.query.UpdateIdentityUserAvatar(ctx, sqlc.UpdateIdentityUserAvatarParams{
		ID:        id,
		AvatarUrl: avatarURL,
		UpdatedBy: id,
	}))
}

func (s *DB) UpdateUserStatus(ctx context.Context, id int64, oldStatus, newStatus entity.UserStatus) (err error) {
	ctx, span := s.startSpan(ctx, "UpdateUserStatus")
	defer func() { s.endSpan(span, err) }()

	return s.mapError(s.query.UpdateIdentityUserStatus(ctx, sqlc.UpdateIdentityUserStatusParams{
		ID:        id,
		NewStatus: newStatus,
		OldStatus: oldStatus,
	}))
}

func (s *DB) UpdateUserCredential(ctx context.Context, userID int64, hash string) (err error) {
	ctx, span := s.startSpan(ctx, "UpdateUserCredential")
	defer func() { s.endSpan(span, err) }()

	err = s.mapError(s.query.UpdateIdentityUserCredential(ctx, sqlc.UpdateIdentityUserCredentialParams{
		Password: hash,
		UserID:   userID,
	}))
	return err
}

func (s *DB) MarkUserDeleted(ctx context.Context, id, byID int64) (err error) {
	ctx, span := s.startSpan(ctx, "MarkUserDeleted")
	defer func() { s.endSpan(span, err) }()

	return s.mapError(s.query.MarkIdentityUserDeleted(ctx, sqlc.MarkIdentityUserDeletedParams{
		DeletedBy: pgtype.Int8{Valid: true, Int64: byID},
		ID:        id,
	}))
}
