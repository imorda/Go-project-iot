package usecase

import (
	"context"
	"homework/internal/domain"
)

type User struct {
	// TODO добавьте реализацию
}

func NewUser(ur UserRepository, sor SensorOwnerRepository, sr SensorRepository) *User {
	return &User{}
}

func (u *User) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	// TODO добавьте реализацию
	return nil, nil
}

func (u *User) AttachSensorToUser(ctx context.Context, userID, sensorID int64) error {
	// TODO добавьте реализацию
	return nil
}

func (u *User) GetUserSensors(ctx context.Context, userID int64) ([]domain.Sensor, error) {
	// TODO добавьте реализацию
	return nil, nil
}
