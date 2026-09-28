package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Lix-09991/neon-saga-api/internal/models"
)

var (
	ErrUserNotFound  = errors.New("пользователь не найден")
	ErrEmailTaken    = errors.New("email уже занят")
	ErrUsernameTaken = errors.New("username уже занят")
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	query := `
		INSERT INTO users (email, username, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		u.Email, u.Username, u.PasswordHash, string(u.Role),
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			return ErrEmailTaken
		}
		if isUniqueViolation(err, "users_username_key") {
			return ErrUsernameTaken
		}
		return fmt.Errorf("создание пользователя: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `
		SELECT id, email, username, password_hash, avatar, role, created_at, updated_at
		FROM users WHERE email = $1
	`
	return r.scanUser(r.pool.QueryRow(ctx, query, email))
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	const query = `
		SELECT id, email, username, password_hash, avatar, role, created_at, updated_at
		FROM users WHERE username = $1
	`
	return r.scanUser(r.pool.QueryRow(ctx, query, username))
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const query = `
		SELECT id, email, username, password_hash, avatar, role, created_at, updated_at
		FROM users WHERE id = $1
	`
	return r.scanUser(r.pool.QueryRow(ctx, query, id))
}

func (r *UserRepository) scanUser(row pgx.Row) (*models.User, error) {
	var (
		u       models.User
		roleStr string
	)
	err := row.Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash,
		&u.Avatar, &roleStr, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scan пользователя: %w", err)
	}
	u.Role = models.Role(roleStr)
	return &u, nil
}

// isUniqueViolation — проверка на нарушение уникальности (PostgreSQL код 23505)
func isUniqueViolation(err error, constraint string) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "23505") && strings.Contains(s, constraint)
}
