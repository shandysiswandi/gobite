package iam

import (
	"github.com/casbin/casbin/v3"
	"github.com/shandysiswandi/gobite/internal/iam/inbound"
	"github.com/shandysiswandi/gobite/internal/iam/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

type Dependency struct {
	Router   *router.Router
	Enforcer *casbin.Enforcer
}

func New(dep Dependency) error {
	uc := usecase.New(usecase.Dependency{
		Enforcer: dep.Enforcer,
	})

	inbound.RegisterHTTPEndpoint(dep.Router, uc)

	return nil
}
