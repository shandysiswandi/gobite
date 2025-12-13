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
		MfaRequired:  resp.MfaRequired,
		PreAuthToken: resp.PreAuthToken,
	})
}

func (h *HTTPEndpoint) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgrouter.ResponseError(w, err)
		return
	}

	resp, err := h.uc.Register(r.Context(), domain.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		pkgrouter.ResponseError(w, err)
		return
	}

	pkgrouter.Response(w, RegisterResponse{IsNeedVerify: resp.IsNeedVerify})
}

func (h *HTTPEndpoint) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgrouter.ResponseError(w, err)
		return
	}

	resp, err := h.uc.Logout(r.Context(), domain.LogoutInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		pkgrouter.ResponseError(w, err)
		return
	}

	pkgrouter.Response(w, LogoutResponse{Success: resp.Success})
}

func (h *HTTPEndpoint) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgrouter.ResponseError(w, err)
		return
	}

	resp, err := h.uc.RefreshToken(r.Context(), domain.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		pkgrouter.ResponseError(w, err)
		return
	}

	pkgrouter.Response(w, RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}
