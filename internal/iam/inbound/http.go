package inbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

type uc interface {
	ListPolicies(ctx context.Context) ([][]string, error)
	AddPolicies(ctx context.Context, policies [][]string) error
	RemovePolicies(ctx context.Context, policies [][]string) error

	ListGroupingPolicies(ctx context.Context) ([][]string, error)
	AddGroupingPolicies(ctx context.Context, policies [][]string) error
	RemoveGroupingPolicies(ctx context.Context, policies [][]string) error

	ListRoles(ctx context.Context) ([]string, error)
	ListUsersForRole(ctx context.Context, role string) ([]string, error)
	ListRolesForUser(ctx context.Context, user string) ([]string, error)
	ListRolePermissions(ctx context.Context, role string) ([][]string, error)
	ListUserPermissions(ctx context.Context, user string) ([][]string, error)

	Enforce(ctx context.Context, sub, obj, act string) (bool, error)
}

func RegisterHTTPEndpoint(r *router.Router, uc uc) {
	end := &HTTPEndpoint{uc: uc}

	r.GET("/api/v1/iam/policies", end.ListPolicies)
	r.POST("/api/v1/iam/policies", end.AddPolicies)
	r.DELETE("/api/v1/iam/policies", end.RemovePolicies)

	r.GET("/api/v1/iam/grouping-policies", end.ListGroupingPolicies)
	r.POST("/api/v1/iam/grouping-policies", end.AddGroupingPolicies)
	r.DELETE("/api/v1/iam/grouping-policies", end.RemoveGroupingPolicies)

	r.GET("/api/v1/iam/roles", end.ListRoles)
	r.GET("/api/v1/iam/roles/:role/users", end.ListRoleUsers)
	r.GET("/api/v1/iam/users/:user/roles", end.ListUserRoles)

	r.GET("/api/v1/iam/roles/:role/permissions", end.ListRolePermissions)
	r.GET("/api/v1/iam/users/:user/permissions", end.ListUserPermissions)

	r.POST("/api/v1/iam/enforce", end.Enforce)
}
