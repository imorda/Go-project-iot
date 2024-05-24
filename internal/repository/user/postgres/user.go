package postgres

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/usecase"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

const saveUserQuery = `INSERT INTO users (name) VALUES ($1) RETURNING id`

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	row := r.pool.QueryRow(ctx, saveUserQuery, user.Name)

	if err := row.Scan(&user.ID); err != nil {
		return fmt.Errorf("unable to save user to pg: %w", err)
	}

	return nil
}

const getUserByIDQuery = `SELECT id, name FROM users WHERE id = $1`

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, getUserByIDQuery, id)

	user := &domain.User{}
	if err := row.Scan(&user.ID, &user.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrUserNotFound
		}
		return nil, fmt.Errorf("unable to find user by id: %w", err)
	}

	return user, nil
}
