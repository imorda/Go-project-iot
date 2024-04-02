package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"sync"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	storage map[int64]domain.User
	lastId  int64
	mu      sync.Mutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		storage: make(map[int64]domain.User),
		lastId:  0,
	}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if user == nil {
		return errors.New("got nil user at SaveUser()")
	}

	r.mu.Lock()
	r.lastId++
	id := r.lastId

	user.ID = id
	r.storage[id] = *user
	r.mu.Unlock()

	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	val, exists := r.storage[id]
	r.mu.Unlock()
	if !exists {
		return nil, ErrUserNotFound
	}

	return &val, nil
}
