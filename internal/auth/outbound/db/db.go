package db

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

type DB struct {
	conn  *pgxpool.Pool
	query *sqlc.Queries
}

func NewDB(conn *pgxpool.Pool) *DB {
	return &DB{
		conn:  conn,
		query: sqlc.New(conn),
	}
}

func (s *DB) GetUserLoginInfo(ctx context.Context, email string) (*entity.UserLoginInfo, error) {
	result, err := s.query.GetAuthUserLoginInfo(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, goerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &entity.UserLoginInfo{
		ID:       result.ID,
		Email:    result.Email,
		Status:   result.Status,
		Password: result.Password,
		HasMFA:   result.HasMfa,
	}, nil
}

func (s *DB) GetUserCredentialInfo(ctx context.Context, id int64) (*entity.UserCredentialInfo, error) {
	result, err := s.query.GetAuthUserCredentialInfo(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, goerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &entity.UserCredentialInfo{
		ID:       result.ID,
		Status:   result.Status,
		Email:    result.Email,
		Password: result.Password,
	}, nil
}

func (s *DB) CreateRefreshToken(ctx context.Context, in entity.RefreshToken) error {
	return s.query.CreateAuthRefreshToken(ctx, sqlc.CreateAuthRefreshTokenParams{
		ID:        in.ID,
		UserID:    in.UserID,
		Token:     in.Token,
		ExpiresAt: pgtype.Timestamptz{Valid: true, Time: in.ExpiresAt},
		Metadata:  in.Metadata,
	})
}

func (s *DB) CreateChallenge(ctx context.Context, in entity.Challenge) error {
	return s.query.CreateAuthChallenge(ctx, sqlc.CreateAuthChallengeParams{
		ID:        in.ID,
		UserID:    in.UserID,
		Token:     in.Token,
		Purpose:   in.Purpose,
		ExpiresAt: pgtype.Timestamptz{Valid: true, Time: in.ExpiresAt},
		Metadata:  in.Metadata,
	})
}

func (s *DB) GetChallengeUserByTokenPurpose(ctx context.Context, token string, p entity.ChallengePurpose) (*entity.ChallengeUser, error) {
	result, err := s.query.GetAuthChallengeUserByTokenPurpose(ctx, sqlc.GetAuthChallengeUserByTokenPurposeParams{
		Token:   token,
		Purpose: p,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, goerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &entity.ChallengeUser{
		ChallengeID:       result.ID,
		ChallengePurpose:  result.Purpose,
		ChallengeToken:    result.Token,
		ChallengeMetadata: result.Metadata,
		UserID:            result.UserID,
		UserEmail:         result.Email,
		UserStatus:        result.Status,
	}, nil
}

func (s *DB) DeleteChallenge(ctx context.Context, id int64) error {
	return s.query.DeleteAuthChallengeByID(ctx, id)
}

func (s *DB) RevokeRefreshToken(ctx context.Context, token string) error {
	return s.query.RevokeAuthRefreshToken(ctx, token)
}

func (s *DB) RevokeAllRefreshToken(ctx context.Context, userID int64) error {
	return s.query.RevokeAllAuthRefreshToken(ctx, userID)
}

func (s *DB) GetUserRefreshToken(ctx context.Context, token string) (*entity.UserRefreshToken, error) {
	result, err := s.query.GetAuthUserRefreshToken(ctx, token)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, goerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var replacedByTokenID *int64
	if result.ReplacedByTokenID.Valid {
		replacedByTokenID = &result.ReplacedByTokenID.Int64
	}

	return &entity.UserRefreshToken{
		UserID:                   result.UserID,
		UserEmail:                result.Email,
		UserStatus:               result.UserStatus,
		RefreshID:                result.ID,
		RefreshToken:             result.Token,
		RefreshRevoked:           result.Revoked,
		RefreshReplacedByTokenID: replacedByTokenID,
		RefreshExpiresAt:         result.ExpiresAt.Time,
	}, nil
}

func (s *DB) RotateRefreshToken(ctx context.Context, ro entity.RotateRefreshToken) error {
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

	if err := wtx.ReplaceAuthRefreshToken(ctx, sqlc.ReplaceAuthRefreshTokenParams{
		NewTokenID: ro.NewID,
		OldTokenID: ro.OldID,
	}); err != nil {
		return err
	}

	if err := wtx.CreateAuthRefreshToken(ctx, sqlc.CreateAuthRefreshTokenParams{
		ID:        ro.NewID,
		UserID:    ro.UserID,
		Token:     ro.NewToken,
		ExpiresAt: pgtype.Timestamptz{Valid: true, Time: ro.NewExpiresAt},
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *DB) GetUserByEmail(ctx context.Context, email string, includeDeleted bool) (*entity.User, error) {
	var result sqlc.AuthUser
	var err error

	if includeDeleted {
		result, err = s.query.GetAuthUserByEmailIncludeDeleted(ctx, email)
	} else {
		result, err = s.query.GetAuthUserByEmail(ctx, email)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, goerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	item := &entity.User{
		ID:        result.ID,
		Email:     result.Email,
		FullName:  result.FullName,
		AvatarURL: result.AvatarUrl,
		Status:    result.Status,
	}

	if result.CreatedAt.Valid {
		item.CreatedAt = result.CreatedAt.Time
	}
	if result.UpdatedAt.Valid {
		item.UpdatedAt = result.UpdatedAt.Time
	}

	return item, nil
}

func (s *DB) CreateMFAFactorAndChallenge(ctx context.Context, factor entity.MFAFactor, challe entity.Challenge) error {
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

	if err := wtx.CreateAuthMFAFactor(ctx, sqlc.CreateAuthMFAFactorParams{
		ID:           factor.ID,
		UserID:       factor.UserID,
		Type:         factor.Type,
		FriendlyName: factor.FriendlyName,
		Secret:       factor.Secret,
		KeyVersion:   factor.KeyVersion,
		IsVerified:   factor.IsVerified,
	}); err != nil {
		return err
	}

	if err := wtx.CreateAuthChallenge(ctx, sqlc.CreateAuthChallengeParams{
		ID:        challe.ID,
		UserID:    challe.UserID,
		Token:     challe.Token,
		Purpose:   challe.Purpose,
		ExpiresAt: pgtype.Timestamptz{Valid: true, Time: challe.ExpiresAt},
		Metadata:  challe.Metadata,
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *DB) GetMFAFactorByUserID(ctx context.Context, userID int64, isVerified bool) ([]entity.MFAFactor, error) {
	items, err := s.query.GetAuthMFAFactorByUserID(ctx, sqlc.GetAuthMFAFactorByUserIDParams{
		UserID:     userID,
		IsVerified: isVerified,
	})
	if err != nil {
		return nil, err
	}

	result := make([]entity.MFAFactor, 0, len(items))
	for _, item := range items {
		m := entity.MFAFactor{
			ID:           item.ID,
			UserID:       item.UserID,
			Type:         item.Type,
			FriendlyName: item.FriendlyName,
			Secret:       item.Secret,
			KeyVersion:   item.KeyVersion,
			IsVerified:   item.IsVerified,
		}

		if item.CreatedAt.Valid {
			m.CreatedAt = item.CreatedAt.Time
		}
		if item.UpdatedAt.Valid {
			m.UpdatedAt = item.UpdatedAt.Time
		}

		result = append(result, m)
	}

	return result, nil
}

func (s *DB) UpdateUserProfile(ctx context.Context, id int64, fullName string) error {
	return s.query.UpdateAuthUserName(ctx, sqlc.UpdateAuthUserNameParams{
		ID:       id,
		FullName: fullName,
	})
}

func (s *DB) UserRegistration(ctx context.Context, user entity.User, hash string) error {
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

	if err := wtx.CreateAuthUser(ctx, sqlc.CreateAuthUserParams{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarUrl: user.AvatarURL,
		Status:    user.Status,
	}); err != nil {
		return err
	}

	if err := wtx.CreateAuthUserCredential(ctx, sqlc.CreateAuthUserCredentialParams{
		UserID:   user.ID,
		Password: hash,
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *DB) UpdateUserCredential(ctx context.Context, userID int64, hash string) error {
	return s.query.UpdateAuthUserCredential(ctx, sqlc.UpdateAuthUserCredentialParams{
		Password: hash,
		UserID:   userID,
	})
}
