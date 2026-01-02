package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type (
	ProfileSettingMFAOutput struct {
		TOTPEnabled       bool
		BackupCodeEnabled bool
		SMSEnabled        bool
	}
)

func (s *Usecase) ProfileSettingMFA(ctx context.Context) (*ProfileSettingMFAOutput, error) {
	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	mfas, err := s.repoDB.GetMFAFactorByUserID(ctx, clm.UserID, true)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get mfa factor by user_id", "user_id", clm.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	resp := ProfileSettingMFAOutput{}
	for _, mfa := range mfas {
		switch mfa.Type {
		case entity.MFATypeBackupCode:
			resp.BackupCodeEnabled = true
		case entity.MFATypeSMS:
			resp.SMSEnabled = false
		case entity.MFATypeTOTP:
			resp.TOTPEnabled = true
		}
	}

	return &resp, nil
}
