package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/Lix-09991/neon-saga-api/internal/models"
	"github.com/Lix-09991/neon-saga-api/internal/repository"
	"github.com/Lix-09991/neon-saga-api/pkg/hash"
)

var (
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	ErrInvalidToken       = errors.New("неверный токен")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret []byte
	jwtTTL    time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret, jwtTTL string) *AuthService {
	ttl := 7 * 24 * time.Hour
	if d, err := time.ParseDuration(jwtTTL); err == nil {
		ttl = d
	}
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		jwtTTL:    ttl,
	}
}

type RegisterInput struct {
	Username string
	Password string
}

type LoginInput struct {
	Username string
	Password string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*models.User, string, error) {
	hashed, err := hash.HashPassword(in.Password)
	if err != nil {
		return nil, "", fmt.Errorf("хеширование пароля: %w", err)
	}

	u := &models.User{
		Username:     in.Username,
		PasswordHash: hashed,
		Role:         models.RoleUser,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, "", err
	}

	token, err := s.generateToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*models.User, string, error) {
	u, err := s.userRepo.FindByUsername(ctx, in.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if !hash.CheckPassword(in.Password, u.PasswordHash) {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.generateToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *AuthService) generateToken(u *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      u.ID.String(),
		"username": u.Username,
		"role":     string(u.Role),
		"exp":      time.Now().Add(s.jwtTTL).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
