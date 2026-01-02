package inbound

import (
	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

type HTTPEndpoint struct {
	uc uc
}

// @Summary Get my permissions
// @Tags IAM
// @Success 200 {object} router.successResponse{data=MePermissionsResponse} "Permissions list"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/iam/me/permissions [get]
func (h *HTTPEndpoint) MePermissions(r *router.Request) (any, error) {
	resp, err := h.uc.MePermissions(r.Context())
	if err != nil {
		return nil, err
	}

	if resp == nil {
		resp = map[string][]string{}
	}

	return MePermissionsResponse{Permissions: resp}, nil
}
