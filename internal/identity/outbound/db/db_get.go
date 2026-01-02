package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
)

func (s *DB) GetUserLoginInfo(ctx context.Context, email string) (_ *entity.UserLoginInfo, err error) {
	ctx, span := s.startSpan(ctx, "GetUserLoginInfo")
	defer func() { s.endSpan(span, err) }()

	result, err := s.query.GetIdentityUserLoginInfo(ctx, email)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &entity.UserLoginInfo{
		ID:       result.ID,
		Email:    result.Email,
		Status:   result.Status,
		Password: result.Password,
		HasMFA:   result.HasMfa,
	}, nil
}

func (s *DB) GetUserCredentialInfo(ctx context.Context, id int64) (_ *entity.UserCredentialInfo, err error) {
	ctx, span := s.startSpan(ctx, "GetUserCredentialInfo")
	defer func() { s.endSpan(span, err) }()

	result, err := s.query.GetIdentityUserCredentialInfo(ctx, id)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &entity.UserCredentialInfo{
		ID:       result.ID,
		Status:   result.Status,
		Email:    result.Email,
		Password: result.Password,
	}, nil
}

func (s *DB) GetChallengeUserByTokenPurpose(ctx context.Context, token string, p entity.ChallengePurpose) (_ *entity.ChallengeUser, err error) {
	ctx, span := s.startSpan(ctx, "GetChallengeUserByTokenPurpose")
	defer func() { s.endSpan(span, err) }()

	result, err := s.query.GetIdentityChallengeUserByTokenPurpose(ctx, sqlc.GetIdentityChallengeUserByTokenPurposeParams{
		Token:   token,
		Purpose: p,
	})
	if err != nil {
		return nil, s.mapError(err)
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

func (s *DB) GetUserRefreshToken(ctx context.Context, token string) (_ *entity.UserRefreshToken, err error) {
	ctx, span := s.startSpan(ctx, "GetUserRefreshToken")
	defer func() { s.endSpan(span, err) }()

	result, err := s.query.GetIdentityUserRefreshToken(ctx, token)
	if err != nil {
		return nil, s.mapError(err)
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

func (s *DB) GetUserByEmail(ctx context.Context, email string, includeDeleted bool) (_ *entity.User, err error) {
	ctx, span := s.startSpan(ctx, "GetUserByEmail")
	defer func() { s.endSpan(span, err) }()

	if includeDeleted {
		result, err := s.query.GetIdentityUserByEmailIncludeDeleted(ctx, email)
		if err != nil {
			return nil, s.mapError(err)
		}

		return &entity.User{
			ID:        result.ID,
			Email:     result.Email,
			FullName:  result.FullName,
			AvatarURL: result.AvatarUrl,
			Status:    result.Status,
		}, nil
	}

	result, err := s.query.GetIdentityUserByEmail(ctx, email)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &entity.User{
		ID:        result.ID,
		Email:     result.Email,
		FullName:  result.FullName,
		AvatarURL: result.AvatarUrl,
		Status:    result.Status,
	}, nil
}

func (s *DB) GetMFAFactorByUserID(ctx context.Context, userID int64, isVerified bool) (_ []entity.MFAFactor, err error) {
	ctx, span := s.startSpan(ctx, "GetMFAFactorByUserID")
	defer func() { s.endSpan(span, err) }()

	items, err := s.query.GetIdentityMFAFactorByUserID(ctx, sqlc.GetIdentityMFAFactorByUserIDParams{
		UserID:     userID,
		IsVerified: isVerified,
	})
	if err != nil {
		return nil, s.mapError(err)
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

		result = append(result, m)
	}

	return result, nil
}

func (s *DB) GetMFAFactorByID(ctx context.Context, id int64, userID int64) (_ *entity.MFAFactor, err error) {
	ctx, span := s.startSpan(ctx, "GetMFAFactorByID")
	defer func() { s.endSpan(span, err) }()

	result, err := s.query.GetIdentityMFAFactorByID(ctx, sqlc.GetIdentityMFAFactorByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, s.mapError(err)
	}

	item := &entity.MFAFactor{
		ID:           result.ID,
		UserID:       result.UserID,
		Type:         result.Type,
		FriendlyName: result.FriendlyName,
		Secret:       result.Secret,
		KeyVersion:   result.KeyVersion,
		IsVerified:   result.IsVerified,
	}

	return item, nil
}

func (s *DB) GetMFABackupCodeByUserID(ctx context.Context, userID int64) (_ []entity.MFABackupCode, err error) {
	ctx, span := s.startSpan(ctx, "GetMFABackupCodeByUserID")
	defer func() { s.endSpan(span, err) }()

	results, err := s.query.GetIdentityMFABackupCodeByUserID(ctx, userID)
	if err != nil {
		return nil, s.mapError(err)
	}

	items := make([]entity.MFABackupCode, 0)
	for _, result := range results {
		items = append(items, entity.MFABackupCode{
			ID:     result.ID,
			UserID: result.UserID,
			Code:   result.Code,
		})
	}

	return items, nil
}

func (s *DB) GetUserList(ctx context.Context, filter entity.UserListFilterData) (_ []entity.User, _ int64, err error) {
	ctx, span := s.startSpan(ctx, "GetUserList")
	defer func() { s.endSpan(span, err) }()

	dateFrom := pgtype.Timestamptz{Time: filter.DateFrom, Valid: !filter.DateFrom.IsZero()}
	dateTo := pgtype.Timestamptz{Time: filter.DateTo, Valid: !filter.DateTo.IsZero()}

	items, err := s.query.GetIdentityUserFilter(ctx, sqlc.GetIdentityUserFilterParams{
		FilterByStatus:   filter.IsFilterByStatus,
		FilterBySearch:   filter.IsFilterBySearch,
		FilterByDateFrom: !filter.DateFrom.IsZero(),
		FilterByDateTo:   !filter.DateTo.IsZero(),
		Statuses:         filter.Statuses,
		Search:           filter.Search,
		DateFrom:         dateFrom,
		DateTo:           dateTo,
		PageOffset:       filter.Page,
		PageLimit:        filter.Size,
		OrderBy:          fmt.Sprintf("%s:%s", filter.OrderBy, filter.OrderDirection),
	})
	if err != nil {
		return nil, 0, s.mapError(err)
	}

	fmt.Println("LL", len(items))

	users := make([]entity.User, 0, len(items))
	for _, item := range items {
		user := entity.User{
			ID:        item.ID,
			Email:     item.Email,
			FullName:  item.FullName,
			AvatarURL: item.AvatarUrl,
			Status:    item.Status,
		}
		if item.UpdatedAt.Valid {
			user.UpdatedAt = item.UpdatedAt.Time
		}

		users = append(users, user)
	}

	count, err := s.query.CountIdentityUserFilter(ctx, sqlc.CountIdentityUserFilterParams{
		FilterByStatus:   filter.IsFilterByStatus,
		FilterBySearch:   filter.IsFilterBySearch,
		FilterByDateFrom: !filter.DateFrom.IsZero(),
		FilterByDateTo:   !filter.DateTo.IsZero(),
		Statuses:         filter.Statuses,
		Search:           filter.Search,
		DateFrom:         dateFrom,
		DateTo:           dateTo,
	})
	if err != nil {
		return nil, 0, s.mapError(err)
	}

	return users, count, nil
}

func (s *DB) GetUserByID(ctx context.Context, id int64, includeDeleted bool) (_ *entity.User, err error) {
	ctx, span := s.startSpan(ctx, "GetUserByID")
	defer func() { s.endSpan(span, err) }()

	if includeDeleted {
		result, err := s.query.GetIdentityUserByIDIncludeDeleted(ctx, id)
		if err != nil {
			return nil, s.mapError(err)
		}

		item := &entity.User{
			ID:        result.ID,
			Email:     result.Email,
			FullName:  result.FullName,
			AvatarURL: result.AvatarUrl,
			Status:    result.Status,
		}
		if result.UpdatedAt.Valid {
			item.UpdatedAt = result.UpdatedAt.Time
		}
		if result.DeletedAt.Valid {
			item.DeletedAt = &result.DeletedAt.Time
		}

		return item, nil
	}

	result, err := s.query.GetIdentityUserByID(ctx, id)
	if err != nil {
		return nil, s.mapError(err)
	}

	item := &entity.User{
		ID:        result.ID,
		Email:     result.Email,
		FullName:  result.FullName,
		AvatarURL: result.AvatarUrl,
		Status:    result.Status,
	}
	if result.UpdatedAt.Valid {
		item.UpdatedAt = result.UpdatedAt.Time
	}
	if result.DeletedAt.Valid {
		item.DeletedAt = &result.DeletedAt.Time
	}

	return item, nil
}
