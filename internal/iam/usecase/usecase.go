package usecase

import (
	"context"
	"strconv"

	"github.com/casbin/casbin/v3"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type Dependency struct {
	Enforcer *casbin.Enforcer
}

type Usecase struct {
	enforcer *casbin.Enforcer
}

func New(dep Dependency) *Usecase {
	return &Usecase{enforcer: dep.Enforcer}
}

func (u *Usecase) MePermissions(ctx context.Context) (map[string][]string, error) {
	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	policies, err := u.enforcer.GetImplicitPermissionsForUser(strconv.FormatInt(clm.UserID, 10))
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
