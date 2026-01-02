package db

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

func (s *DB) NewRegistration(ctx context.Context, user entity.User, chal entity.Challenge, hash string) (err error) {
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

func (s *DB) NewMFAFactorTOTPBackup(ctx context.Context, fTOTP, fBackup entity.MFAFactor, codes []entity.MFABackupCode, challengeID int64) (err error) {
	ctx, span := s.startSpan(ctx, "NewMFAFactorTOTPBackup")
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

	if err := wtx.CreateIdentityMFAFactor(ctx, sqlc.CreateIdentityMFAFactorParams{
		ID:           fBackup.ID,
		UserID:       fBackup.UserID,
		Type:         fBackup.Type,
		FriendlyName: fBackup.FriendlyName,
		Secret:       fBackup.Secret,
		KeyVersion:   fBackup.KeyVersion,
		IsVerified:   fBackup.IsVerified,
	}); err != nil {
		return s.mapError(err)
	}

	if err := wtx.DeleteIdentityChallengeByID(ctx, challengeID); err != nil {
		return s.mapError(err)
	}

	if err := wtx.DeleteIdentityMFABackupCodeByUserID(ctx, codes[0].UserID); err != nil {
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

func (s *DB) NewBackupCodes(ctx context.Context, codes []entity.MFABackupCode) (err error) {
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

	if err := wtx.DeleteIdentityMFABackupCodeByUserID(ctx, codes[0].UserID); err != nil {
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
