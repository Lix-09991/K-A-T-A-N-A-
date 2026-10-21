package handler

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	custommw "github.com/Lix-09991/neon-saga-api/internal/middleware"
	"github.com/Lix-09991/neon-saga-api/internal/repository"
	"github.com/Lix-09991/neon-saga-api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	validate    *validator.Validate
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService, validate: validator.New()}
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name: "token", Value: token, Path: "/",
		HttpOnly: true, Secure: false, SameSite: http.SameSiteLaxMode,
		MaxAge: 7 * 24 * 60 * 60,
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: "token", Value: "", Path: "/",
		HttpOnly: true, Secure: false, SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})
}

type registerRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

type loginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type authResponse struct {
	User  interface{} `json:"user"`
	Token string      `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}

	user, token, err := h.authService.Register(r.Context(), service.RegisterInput{
		Username: req.Username, Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, repository.ErrUsernameTaken) {
			writeError(w, http.StatusConflict, "username уже занят")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка регистрации")
		return
	}

	setAuthCookie(w, token)
	writeJSON(w, http.StatusCreated, authResponse{User: user, Token: token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}

	user, token, err := h.authService.Login(r.Context(), service.LoginInput{
		Username: req.Username, Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "неверный логин или пароль")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка входа")
		return
	}

	setAuthCookie(w, token)
	writeJSON(w, http.StatusOK, authResponse{User: user, Token: token})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "вы вышли"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userFromCtx, ok := custommw.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "не авторизован")
		return
	}
	user, err := h.authService.GetUserByID(r.Context(), userFromCtx.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения профиля")
		return
	}
	writeJSON(w, http.StatusOK, user)
}
