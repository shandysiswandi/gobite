package inbound

import (
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

type HTTPEndpoint struct {
	uc uc
}

// ListPolicies returns all policy rules (p).
// @Summary List policies
// @Description Returns all policy rules in the enforcer.
// @Tags IAM
// @Produce json
// @Success 200 {object} router.successResponse{data=PoliciesResponse} "Policy list"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/policies [get]
func (h *HTTPEndpoint) ListPolicies(r *router.Request) (any, error) {
	policies, err := h.uc.ListPolicies(r.Context())
	if err != nil {
		return nil, err
	}

	return PoliciesResponse{Policies: toPolicyResponses(policies)}, nil
}

// AddPolicies adds one or more policy rules (p).
// @Summary Add policies
// @Description Adds policy rules in single or batch format.
// @Tags IAM
// @Accept json
// @Produce json
// @Param request body PoliciesRequest true "Policies payload"
// @Success 200 {object} router.successResponse{data=MutationResponse} "Policies added"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/policies [post]
func (h *HTTPEndpoint) AddPolicies(r *router.Request) (any, error) {
	var req PoliciesRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	policies, err := policiesFromRequest(req)
	if err != nil {
		return nil, err
	}

	return MutationResponse{}, h.uc.AddPolicies(r.Context(), policies)
}

// RemovePolicies removes one or more policy rules (p).
// @Summary Remove policies
// @Description Removes policy rules in single or batch format.
// @Tags IAM
// @Accept json
// @Produce json
// @Param request body PoliciesRequest true "Policies payload"
// @Success 200 {object} router.successResponse{data=MutationResponse} "Policies removed"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/policies [delete]
func (h *HTTPEndpoint) RemovePolicies(r *router.Request) (any, error) {
	var req PoliciesRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	policies, err := policiesFromRequest(req)
	if err != nil {
		return nil, err
	}

	return MutationResponse{}, h.uc.RemovePolicies(r.Context(), policies)
}

// ListGroupingPolicies returns all grouping rules (g).
// @Summary List grouping policies
// @Description Returns all grouping policies in the enforcer.
// @Tags IAM
// @Produce json
// @Success 200 {object} router.successResponse{data=GroupingPoliciesResponse} "Grouping policy list"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/grouping-policies [get]
func (h *HTTPEndpoint) ListGroupingPolicies(r *router.Request) (any, error) {
	policies, err := h.uc.ListGroupingPolicies(r.Context())
	if err != nil {
		return nil, err
	}

	return GroupingPoliciesResponse{Policies: toGroupingPolicyResponses(policies)}, nil
}

// AddGroupingPolicies adds one or more grouping rules (g).
// @Summary Add grouping policies
// @Description Adds grouping policies in single or batch format.
// @Tags IAM
// @Accept json
// @Produce json
// @Param request body GroupingPoliciesRequest true "Grouping policies payload"
// @Success 200 {object} router.successResponse{data=MutationResponse} "Grouping policies added"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/grouping-policies [post]
func (h *HTTPEndpoint) AddGroupingPolicies(r *router.Request) (any, error) {
	var req GroupingPoliciesRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	policies, err := groupingPoliciesFromRequest(req)
	if err != nil {
		return nil, err
	}

	return MutationResponse{}, h.uc.AddGroupingPolicies(r.Context(), policies)
}

// RemoveGroupingPolicies removes one or more grouping rules (g).
// @Summary Remove grouping policies
// @Description Removes grouping policies in single or batch format.
// @Tags IAM
// @Accept json
// @Produce json
// @Param request body GroupingPoliciesRequest true "Grouping policies payload"
// @Success 200 {object} router.successResponse{data=MutationResponse} "Grouping policies removed"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/grouping-policies [delete]
func (h *HTTPEndpoint) RemoveGroupingPolicies(r *router.Request) (any, error) {
	var req GroupingPoliciesRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	policies, err := groupingPoliciesFromRequest(req)
	if err != nil {
		return nil, err
	}

	return MutationResponse{}, h.uc.RemoveGroupingPolicies(r.Context(), policies)
}

// ListRoles returns all roles from grouping policies.
// @Summary List roles
// @Description Returns all roles in the enforcer.
// @Tags IAM
// @Produce json
// @Success 200 {object} router.successResponse{data=RolesResponse} "Role list"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/roles [get]
func (h *HTTPEndpoint) ListRoles(r *router.Request) (any, error) {
	roles, err := h.uc.ListRoles(r.Context())
	if err != nil {
		return nil, err
	}

	return RolesResponse{Roles: roles}, nil
}

// ListRoleUsers returns users assigned to a role.
// @Summary List role users
// @Description Returns all users assigned to a role.
// @Tags IAM
// @Produce json
// @Param role path string true "Role name"
// @Success 200 {object} router.successResponse{data=RoleUsersResponse} "Role users list"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/roles/{role}/users [get]
func (h *HTTPEndpoint) ListRoleUsers(r *router.Request) (any, error) {
	role := r.GetParam("role")
	users, err := h.uc.ListUsersForRole(r.Context(), role)
	if err != nil {
		return nil, err
	}

	return RoleUsersResponse{Role: role, Users: users}, nil
}

// ListUserRoles returns roles assigned to a user.
// @Summary List user roles
// @Description Returns all roles assigned to a user.
// @Tags IAM
// @Produce json
// @Param user path string true "User identifier"
// @Success 200 {object} router.successResponse{data=UserRolesResponse} "User roles list"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/users/{user}/roles [get]
func (h *HTTPEndpoint) ListUserRoles(r *router.Request) (any, error) {
	user := r.GetParam("user")
	roles, err := h.uc.ListRolesForUser(r.Context(), user)
	if err != nil {
		return nil, err
	}

	return UserRolesResponse{User: user, Roles: roles}, nil
}

// ListRolePermissions returns permissions for a role.
// @Summary List role permissions
// @Description Returns permissions granted to a role.
// @Tags IAM
// @Produce json
// @Param role path string true "Role name"
// @Success 200 {object} router.successResponse{data=RolePermissionsResponse} "Role permissions list"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/roles/{role}/permissions [get]
func (h *HTTPEndpoint) ListRolePermissions(r *router.Request) (any, error) {
	role := r.GetParam("role")
	perms, err := h.uc.ListRolePermissions(r.Context(), role)
	if err != nil {
		return nil, err
	}

	return RolePermissionsResponse{Role: role, Permissions: toPermissionResponses(perms)}, nil
}

// ListUserPermissions returns permissions for a user.
// @Summary List user permissions
// @Description Returns permissions granted to a user (implicit).
// @Tags IAM
// @Produce json
// @Param user path string true "User identifier"
// @Success 200 {object} router.successResponse{data=UserPermissionsResponse} "User permissions list"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/users/{user}/permissions [get]
func (h *HTTPEndpoint) ListUserPermissions(r *router.Request) (any, error) {
	user := r.GetParam("user")
	perms, err := h.uc.ListUserPermissions(r.Context(), user)
	if err != nil {
		return nil, err
	}

	return UserPermissionsResponse{User: user, Permissions: toPermissionResponses(perms)}, nil
}

// Enforce checks whether a request is allowed.
// @Summary Enforce policy
// @Description Evaluates a policy for the given subject, object, and action.
// @Tags IAM
// @Accept json
// @Produce json
// @Param request body EnforceRequest true "Enforcement payload"
// @Success 200 {object} router.successResponse{data=EnforceResponse} "Enforcement result"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/enforce [post]
func (h *HTTPEndpoint) Enforce(r *router.Request) (any, error) {
	var req EnforceRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	allowed, err := h.uc.Enforce(r.Context(), req.Sub, req.Obj, req.Act)
	if err != nil {
		return nil, err
	}

	return EnforceResponse{Allowed: allowed}, nil
}

func policiesFromRequest(req PoliciesRequest) ([][]string, error) {
	var entries []PolicyRequest
	if req.Policy != nil {
		entries = append(entries, *req.Policy)
	}
	if len(req.Policies) > 0 {
		entries = append(entries, req.Policies...)
	}
	if len(entries) == 0 {
		return nil, goerror.NewInvalidInput(nil, "policies", "policies is required")
	}

	policies := make([][]string, 0, len(entries))
	for _, p := range entries {
		if p.Sub == "" || p.Obj == "" || p.Act == "" {
			return nil, goerror.NewInvalidInput(nil, "sub", "sub is required", "obj", "obj is required", "act", "act is required")
		}
		policies = append(policies, []string{p.Sub, p.Obj, p.Act})
	}

	return policies, nil
}

func groupingPoliciesFromRequest(req GroupingPoliciesRequest) ([][]string, error) {
	var entries []GroupingPolicyRequest
	if req.Policy != nil {
		entries = append(entries, *req.Policy)
	}
	if len(req.Policies) > 0 {
		entries = append(entries, req.Policies...)
	}
	if len(entries) == 0 {
		return nil, goerror.NewInvalidInput(nil, "policies", "policies is required")
	}

	policies := make([][]string, 0, len(entries))
	for _, p := range entries {
		if p.User == "" || p.Role == "" {
			return nil, goerror.NewInvalidInput(nil, "user", "user is required", "role", "role is required")
		}
		policies = append(policies, []string{p.User, p.Role})
	}

	return policies, nil
}

func toPolicyResponses(policies [][]string) []PolicyResponse {
	resp := make([]PolicyResponse, 0, len(policies))
	for _, p := range policies {
		if len(p) < 3 {
			continue
		}
		resp = append(resp, PolicyResponse{Sub: p[0], Obj: p[1], Act: p[2]})
	}
	return resp
}

func toGroupingPolicyResponses(policies [][]string) []GroupingPolicyResponse {
	resp := make([]GroupingPolicyResponse, 0, len(policies))
	for _, p := range policies {
		if len(p) < 2 {
			continue
		}
		resp = append(resp, GroupingPolicyResponse{User: p[0], Role: p[1]})
	}
	return resp
}

func toPermissionResponses(policies [][]string) []PermissionResponse {
	resp := make([]PermissionResponse, 0, len(policies))
	for _, p := range policies {
		if len(p) < 3 {
			continue
		}
		resp = append(resp, PermissionResponse{Sub: p[0], Obj: p[1], Act: p[2]})
	}
	return resp
}
