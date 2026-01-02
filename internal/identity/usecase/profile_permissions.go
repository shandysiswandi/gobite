package usecase

import (
	"context"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

func (s *Usecase) ProfilePermissions(ctx context.Context) (map[string][]string, error) {
	ctx, span := s.startSpan(ctx, "ProfilePermissions")
	defer span.End()

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	policies, err := s.enforcer.GetImplicitPermissionsForUser(strconv.FormatInt(clm.UserID, 10))
	if err != nil {
		return nil, err
	}

	permissions := make(map[string][]string)
	for _, policy := range policies {
		if len(policy) < 3 {
			// Skip malformed policies missing subject/object/action.
			continue
		}

		permissions[policy[1]] = append(permissions[policy[1]], policy[2])
	}

	return permissions, nil
}
