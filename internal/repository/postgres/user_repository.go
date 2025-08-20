package postgres

import (
	"context"
	"fmt"

	"gofemart/internal/database"
	"gofemart/internal/models"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User, passwordHash string) error {
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (login, password_hash) 
		VALUES ($1, $2) 
		RETURNING id, login, created_at, updated_at`,
		user.Login, passwordHash).Scan(&user.ID, &user.Login, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx, `
		SELECT id, login, password_hash, created_at, updated_at 
		FROM users WHERE login = $1`, login).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE login = $1)", login).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}
	return exists, nil
}
