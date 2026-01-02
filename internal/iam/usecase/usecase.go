package usecase

import (
	"context"

	"github.com/casbin/casbin/v3"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
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

func (u *Usecase) ListPolicies(ctx context.Context) ([][]string, error) {
	policies, err := u.enforcer.GetPolicy()
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (u *Usecase) AddPolicies(ctx context.Context, policies [][]string) error {
	if len(policies) == 0 {
		return goerror.NewInvalidInput(nil, "policies", "policies is required")
	}

	if len(policies) == 1 {
		_, err := u.enforcer.AddPolicy(policyToInterfaces(policies[0])...)
		return err
	}

	_, err := u.enforcer.AddPolicies(policies)
	return err
}

func (u *Usecase) RemovePolicies(ctx context.Context, policies [][]string) error {
	if len(policies) == 0 {
		return goerror.NewInvalidInput(nil, "policies", "policies is required")
	}

	if len(policies) == 1 {
		_, err := u.enforcer.RemovePolicy(policyToInterfaces(policies[0])...)
		return err
	}

	_, err := u.enforcer.RemovePolicies(policies)
	return err
}

func (u *Usecase) ListGroupingPolicies(ctx context.Context) ([][]string, error) {
	policies, err := u.enforcer.GetGroupingPolicy()
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (u *Usecase) AddGroupingPolicies(ctx context.Context, policies [][]string) error {
	if len(policies) == 0 {
		return goerror.NewInvalidInput(nil, "policies", "policies is required")
	}

	if len(policies) == 1 {
		_, err := u.enforcer.AddGroupingPolicy(policyToInterfaces(policies[0])...)
		return err
	}

	_, err := u.enforcer.AddGroupingPolicies(policies)
	return err
}

func (u *Usecase) RemoveGroupingPolicies(ctx context.Context, policies [][]string) error {
	if len(policies) == 0 {
		return goerror.NewInvalidInput(nil, "policies", "policies is required")
	}

	if len(policies) == 1 {
		_, err := u.enforcer.RemoveGroupingPolicy(policyToInterfaces(policies[0])...)
		return err
	}

	_, err := u.enforcer.RemoveGroupingPolicies(policies)
	return err
}

func (u *Usecase) ListRoles(ctx context.Context) ([]string, error) {
	return u.enforcer.GetAllRoles()
}

func (u *Usecase) ListUsersForRole(ctx context.Context, role string) ([]string, error) {
	if role == "" {
		return nil, goerror.NewInvalidInput(nil, "role", "role is required")
	}
	return u.enforcer.GetUsersForRole(role)
}

func (u *Usecase) ListRolesForUser(ctx context.Context, user string) ([]string, error) {
	if user == "" {
		return nil, goerror.NewInvalidInput(nil, "user", "user is required")
	}
	return u.enforcer.GetRolesForUser(user)
}

func (u *Usecase) ListRolePermissions(ctx context.Context, role string) ([][]string, error) {
	if role == "" {
		return nil, goerror.NewInvalidInput(nil, "role", "role is required")
	}
	return u.enforcer.GetPermissionsForUser(role)
}

func (u *Usecase) ListUserPermissions(ctx context.Context, user string) ([][]string, error) {
	if user == "" {
		return nil, goerror.NewInvalidInput(nil, "user", "user is required")
	}
	return u.enforcer.GetImplicitPermissionsForUser(user)
}

func (u *Usecase) Enforce(ctx context.Context, sub, obj, act string) (bool, error) {
	if sub == "" || obj == "" || act == "" {
		return false, goerror.NewInvalidInput(nil, "sub", "sub is required", "obj", "obj is required", "act", "act is required")
	}
	return u.enforcer.Enforce(sub, obj, act)
}

func policyToInterfaces(policy []string) []interface{} {
	out := make([]interface{}, 0, len(policy))
	for _, v := range policy {
		out = append(out, v)
	}
	return out
}
