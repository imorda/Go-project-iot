package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	// TODO добавьте реализацию
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	// TODO добавьте реализацию
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	// TODO добавьте реализацию
	return nil, nil
}
