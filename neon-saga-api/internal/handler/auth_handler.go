package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/Lix-09991/neon-saga-api/internal/repository"
	"github.com/Lix-09991/neon-saga-api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	validate    *validator.Validate
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

// ============================================================
// DTO
// ============================================================
type registerRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type authResponse struct {
	User  interface{} `json:"user"`
	Token string      `json:"token"`
}

// ============================================================
// HANDLERS
// ============================================================
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}

	user, token, err := h.authService.Register(r.Context(), service.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEmailTaken):
			writeError(w, http.StatusConflict, "email уже занят")
		case errors.Is(err, repository.ErrUsernameTaken):
			writeError(w, http.StatusConflict, "username уже занят")
		default:
			writeError(w, http.StatusInternalServerError, "ошибка регистрации")
		}
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{User: user, Token: token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !h.decodeAndValidate(w, r, &req) {
		return
	}

	user, token, err := h.authService.Login(r.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "неверный email или пароль")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка входа")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{User: user, Token: token})
}

// ============================================================
// HELPERS
// ============================================================
func (h *AuthHandler) decodeAndValidate(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return false
	}
	if err := h.validate.Struct(dst); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}
