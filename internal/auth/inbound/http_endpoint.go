package inbound

import (
	"encoding/json"
	"net/http"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter"
)

func (h *HTTPEndpoint) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgrouter.ResponseError(w, err)
		return
	}

	resp, err := h.uc.Login(r.Context(), domain.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		pkgrouter.ResponseError(w, err)
		return
	}

	pkgrouter.Response(w, LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}
