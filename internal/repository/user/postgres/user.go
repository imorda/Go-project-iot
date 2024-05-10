package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"homework/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	// TODO добавьте реализацию
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	// TODO добавьте реализацию
	return nil, nil
}
