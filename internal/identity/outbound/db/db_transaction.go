package db

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

func (s *DB) NewRegistration(ctx context.Context, user entity.NewUser, chal entity.Challenge, hash string) (err error) {
	ctx, span := s.startSpan(ctx, "NewRegistration")
	defer func() { s.endSpan(span, err) }()

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if err := wtx.CreateIdentityUser(ctx, sqlc.CreateIdentityUserParams{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarUrl: user.AvatarURL,
		Status:    user.Status,
		CreatedBy: user.CreatedBy,
		UpdatedBy: user.UpdatedBy,
	}); err != nil {
		return s.mapError(err)
	}

	if err := wtx.CreateIdentityUserCredential(ctx, sqlc.CreateIdentityUserCredentialParams{
		UserID:   user.ID,
		Password: hash,
	}); err != nil {
		return s.mapError(err)
	}

	if err := wtx.CreateIdentityChallenge(ctx, sqlc.CreateIdentityChallengeParams{
		ID:        chal.ID,
		UserID:    chal.UserID,
		Token:     chal.Token,
		Purpose:   chal.Purpose,
		ExpiresAt: pgtype.Timestamptz{Valid: true, Time: chal.ExpiresAt},
		Metadata:  chal.Metadata,
	}); err != nil {
		return s.mapError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}

func (s *DB) NewUser(ctx context.Context, user entity.NewUser, hash string) (err error) {
	ctx, span := s.startSpan(ctx, "NewUser")
	defer func() { s.endSpan(span, err) }()

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if err := wtx.CreateIdentityUser(ctx, sqlc.CreateIdentityUserParams{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarUrl: user.AvatarURL,
		Status:    user.Status,
		CreatedBy: user.CreatedBy,
		UpdatedBy: user.UpdatedBy,
	}); err != nil {
		return s.mapError(err)
	}

	if err := wtx.CreateIdentityUserCredential(ctx, sqlc.CreateIdentityUserCredentialParams{
		UserID:   user.ID,
		Password: hash,
	}); err != nil {
		return s.mapError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}

func (s *DB) UpsertUsers(ctx context.Context, users []entity.UpsertUser, hashes map[string]string) (created, updated int, err error) {
	ctx, span := s.startSpan(ctx, "UpsertUsers")
	defer func() { s.endSpan(span, err) }()

	if len(users) == 0 {
		return 0, 0, nil
	}

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	emails := make([]string, 0, len(users))
	for _, user := range users {
		emails = append(emails, user.Email)
	}

	existingUsers, err := wtx.GetIdentityUserByEmailsIncludeDeleted(ctx, emails)
	if err != nil {
		return 0, 0, s.mapError(err)
	}

	existingByEmail := make(map[string]sqlc.GetIdentityUserByEmailsIncludeDeletedRow, len(existingUsers))
	for _, user := range existingUsers {
		existingByEmail[strings.ToLower(user.Email)] = user
	}

	for _, user := range users {
		normalizedEmail := strings.ToLower(user.Email)
		if existing, ok := existingByEmail[normalizedEmail]; ok {
			updated++
			patchArg := sqlc.PatcIdentityUserParams{
				ID:        existing.ID,
				UpdatedBy: pgtype.Int8{Valid: true, Int64: user.UpdatedBy},
			}
			if user.FullName != "" {
				patchArg.FullName = pgtype.Text{Valid: true, String: user.FullName}
				patchArg.AvatarUrl = pgtype.Text{Valid: true, String: user.AvatarURL}

			}
			if user.Status != entity.UserStatusUnknown {
				patchArg.Status = pgtype.Int2{Valid: true, Int16: int16(user.Status)}
			}
			if patchArg.FullName.Valid || patchArg.Status.Valid {
				if err := wtx.PatcIdentityUser(ctx, patchArg); err != nil {
					return 0, 0, s.mapError(err)
				}
			}
			if hash, ok := hashes[normalizedEmail]; ok && hash != "" {
				if err := wtx.UpdateIdentityUserCredential(ctx, sqlc.UpdateIdentityUserCredentialParams{
					UserID:   existing.ID,
					Password: hash,
				}); err != nil {
					return 0, 0, s.mapError(err)
				}
			}
			continue
		}

		created++
		if err := wtx.CreateIdentityUser(ctx, sqlc.CreateIdentityUserParams{
			ID:        user.ID,
			Email:     user.Email,
			FullName:  user.FullName,
			AvatarUrl: user.AvatarURL,
			Status:    user.Status,
			CreatedBy: user.CreatedBy,
			UpdatedBy: user.UpdatedBy,
		}); err != nil {
			return 0, 0, s.mapError(err)
		}

		if hash, ok := hashes[normalizedEmail]; ok && hash != "" {
			if err := wtx.CreateIdentityUserCredential(ctx, sqlc.CreateIdentityUserCredentialParams{
				UserID:   user.ID,
				Password: hash,
			}); err != nil {
				return 0, 0, s.mapError(err)
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, s.mapError(err)
	}

	return created, updated, nil
}

func (s *DB) PatchUser(ctx context.Context, user entity.PatchUser, hash string) (err error) {
	ctx, span := s.startSpan(ctx, "PatchUser")
	defer func() { s.endSpan(span, err) }()

	if hash == "" && user.Email == "" && user.FullName == "" && user.Status.IsUnknown() {
		// nothing to patch
		return nil
	}

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if hash != "" {
		if err := wtx.UpdateIdentityUserCredential(ctx, sqlc.UpdateIdentityUserCredentialParams{
			UserID:   user.ID,
			Password: hash,
		}); err != nil {
			return s.mapError(err)
		}
	}

	patchArg := sqlc.PatcIdentityUserParams{
		ID:        user.ID,
		UpdatedBy: pgtype.Int8{Valid: true, Int64: user.UpdatedBy},
	}
	if user.Email != "" {
		patchArg.Email = pgtype.Text{Valid: true, String: user.Email}
	}
	if user.FullName != "" {
		patchArg.FullName = pgtype.Text{Valid: true, String: user.FullName}
		patchArg.AvatarUrl = pgtype.Text{Valid: true, String: "https://ui-avatars.com/api/?name=" + url.QueryEscape(user.FullName)}
	}
	if !user.Status.IsUnknown() {
		patchArg.Status = pgtype.Int2{Valid: true, Int16: int16(user.Status)}
	}

	if err := wtx.PatcIdentityUser(ctx, patchArg); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}

func (s *DB) NewMFAFactorTOTP(ctx context.Context, fTOTP entity.MFAFactor, challengeID int64) (err error) {
	ctx, span := s.startSpan(ctx, "NewMFAFactorTOTP")
	defer func() { s.endSpan(span, err) }()

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if err := wtx.CreateIdentityMFAFactor(ctx, sqlc.CreateIdentityMFAFactorParams{
		ID:           fTOTP.ID,
		UserID:       fTOTP.UserID,
		Type:         fTOTP.Type,
		FriendlyName: fTOTP.FriendlyName,
		Secret:       fTOTP.Secret,
		KeyVersion:   fTOTP.KeyVersion,
		IsVerified:   fTOTP.IsVerified,
	}); err != nil {
		return s.mapError(err)
	}

	if err := wtx.DeleteIdentityChallengeByID(ctx, challengeID); err != nil {
		return s.mapError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}

func (s *DB) NewRefreshToken(ctx context.Context, ref entity.RefreshToken, challengeID int64) (err error) {
	ctx, span := s.startSpan(ctx, "NewRefreshToken")
	defer func() { s.endSpan(span, err) }()

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if err := wtx.CreateIdentityRefreshToken(ctx, sqlc.CreateIdentityRefreshTokenParams{
		ID:        ref.ID,
		UserID:    ref.UserID,
		Token:     ref.Token,
		ExpiresAt: pgtype.Timestamptz{Valid: true, Time: ref.ExpiresAt},
		Metadata:  ref.Metadata,
	}); err != nil {
		return s.mapError(err)
	}

	if err := wtx.DeleteIdentityChallengeByID(ctx, challengeID); err != nil {
		return s.mapError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}

func (s *DB) NewBackupCodes(ctx context.Context, userID int64, codes []entity.MFABackupCode, factor *entity.MFAFactor) (err error) {
	ctx, span := s.startSpan(ctx, "NewBackupCodes")
	defer func() { s.endSpan(span, err) }()

	if len(codes) == 0 {
		return nil
	}

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if factor != nil {
		if err := wtx.CreateIdentityMFAFactor(ctx, sqlc.CreateIdentityMFAFactorParams{
			ID:           factor.ID,
			UserID:       factor.UserID,
			Type:         factor.Type,
			FriendlyName: factor.FriendlyName,
			Secret:       factor.Secret,
			KeyVersion:   factor.KeyVersion,
			IsVerified:   factor.IsVerified,
		}); err != nil {
			return s.mapError(err)
		}
	}

	if err := wtx.DeleteIdentityMFABackupCodeByUserID(ctx, userID); err != nil {
		return s.mapError(err)
	}

	items := make([]sqlc.CreateIdentityMFABackupCodesParams, 0)
	for i := range codes {
		items = append(items, sqlc.CreateIdentityMFABackupCodesParams{
			ID:     codes[i].ID,
			UserID: codes[i].UserID,
			Code:   codes[i].Code,
		})
	}

	_, err = wtx.CreateIdentityMFABackupCodes(ctx, items)
	if err != nil {
		return s.mapError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}

func (s *DB) VerifyUserRegistration(ctx context.Context, data entity.VerifyUserRegistration) (err error) {
	ctx, span := s.startSpan(ctx, "VerifyUserRegistration")
	defer func() { s.endSpan(span, err) }()

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if err := wtx.UpdateIdentityUserStatus(ctx, sqlc.UpdateIdentityUserStatusParams{
		ID:        data.UserID,
		NewStatus: data.NewUserStatus,
		OldStatus: data.OldUserStatus,
		UpdatedBy: data.UpdatedBy,
	}); err != nil {
		return s.mapError(err)
	}

	if err := wtx.DeleteIdentityChallengeByID(ctx, data.ChallengeID); err != nil {
		return s.mapError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}

func (s *DB) ResetUserPassword(ctx context.Context, userID, challengeID int64, newHash string) (err error) {
	ctx, span := s.startSpan(ctx, "ResetUserPassword")
	defer func() { s.endSpan(span, err) }()

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if err := wtx.UpdateIdentityUserCredential(ctx, sqlc.UpdateIdentityUserCredentialParams{
		Password: newHash,
		UserID:   userID,
	}); err != nil {
		return s.mapError(err)
	}

	if err := wtx.DeleteIdentityChallengeByID(ctx, challengeID); err != nil {
		return s.mapError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}

func (s *DB) VerifyUserMFAFactor(ctx context.Context, userID, challengeID, factorID int64) (err error) {
	ctx, span := s.startSpan(ctx, "VerifyUserMFAFactor")
	defer func() { s.endSpan(span, err) }()

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	if err := wtx.VerifyIdentityMFAFactor(ctx, sqlc.VerifyIdentityMFAFactorParams{
		ID:     factorID,
		UserID: userID,
	}); err != nil {
		return s.mapError(err)
	}

	if err := wtx.DeleteIdentityChallengeByID(ctx, challengeID); err != nil {
		return s.mapError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}

func (s *DB) RotateRefreshToken(ctx context.Context, ro entity.RotateRefreshToken) (err error) {
	ctx, span := s.startSpan(ctx, "RotateRefreshToken")
	defer func() { s.endSpan(span, err) }()

	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rolback", "error", rErr)
		}
	}()

	wtx := s.query.WithTx(tx)

	rows, err := wtx.ReplaceIdentityRefreshToken(ctx, sqlc.ReplaceIdentityRefreshTokenParams{
		NewTokenID: ro.NewID,
		OldTokenID: ro.OldID,
	})
	if err != nil {
		return s.mapError(err)
	}

	if rows == 0 {
		return goerror.ErrNotFound
	}

	if err := wtx.CreateIdentityRefreshToken(ctx, sqlc.CreateIdentityRefreshTokenParams{
		ID:        ro.NewID,
		UserID:    ro.UserID,
		Token:     ro.NewToken,
		ExpiresAt: pgtype.Timestamptz{Valid: true, Time: ro.NewExpiresAt},
	}); err != nil {
		return s.mapError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return s.mapError(err)
	}

	return nil
}
