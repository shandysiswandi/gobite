package inbound

type PolicyRequest struct {
	Sub string `json:"sub"`
	Obj string `json:"obj"`
	Act string `json:"act"`
}

type PoliciesRequest struct {
	Policy   *PolicyRequest  `json:"policy,omitempty"`
	Policies []PolicyRequest `json:"policies,omitempty"`
}

type GroupingPolicyRequest struct {
	User string `json:"user"`
	Role string `json:"role"`
}

type GroupingPoliciesRequest struct {
	Policy   *GroupingPolicyRequest  `json:"policy,omitempty"`
	Policies []GroupingPolicyRequest `json:"policies,omitempty"`
}

type PoliciesResponse struct {
	Policies []PolicyResponse `json:"policies"`
}

type PolicyResponse struct {
	Sub string `json:"sub"`
	Obj string `json:"obj"`
	Act string `json:"act"`
}

type GroupingPoliciesResponse struct {
	Policies []GroupingPolicyResponse `json:"policies"`
}

type GroupingPolicyResponse struct {
	User string `json:"user"`
	Role string `json:"role"`
}

type RolesResponse struct {
	Roles []string `json:"roles"`
}

type RoleUsersResponse struct {
	Role  string   `json:"role"`
	Users []string `json:"users"`
}

type UserRolesResponse struct {
	User  string   `json:"user"`
	Roles []string `json:"roles"`
}

type PermissionResponse struct {
	Sub string `json:"sub"`
	Obj string `json:"obj"`
	Act string `json:"act"`
}

type RolePermissionsResponse struct {
	Role        string               `json:"role"`
	Permissions []PermissionResponse `json:"permissions"`
}

type UserPermissionsResponse struct {
	User        string               `json:"user"`
	Permissions []PermissionResponse `json:"permissions"`
}

type EnforceRequest struct {
	Sub string `json:"sub"`
	Obj string `json:"obj"`
	Act string `json:"act"`
}

type EnforceResponse struct {
	Allowed bool `json:"allowed"`
}

type MutationResponse struct{}
