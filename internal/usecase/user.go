package usecase

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
)

type User struct {
	userRepository        UserRepository
	sensorOwnerRepository SensorOwnerRepository
	sensorRepository      SensorRepository
}

func NewUser(ur UserRepository, sor SensorOwnerRepository, sr SensorRepository) *User {
	return &User{
		userRepository:        ur,
		sensorOwnerRepository: sor,
		sensorRepository:      sr,
	}
}

func (u *User) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if user == nil {
		return user, errors.New("got nil user at RegisterUser()")
	}
	if user.Name == "" {
		return user, ErrInvalidUserName
	}

	if err := u.userRepository.SaveUser(ctx, user); err != nil { // Modifies user assigning new id
		return user, err
	}
	return user, nil
}

func (u *User) AttachSensorToUser(ctx context.Context, userID, sensorID int64) error {
	_, err := u.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("got invalid user id (%v): %w", userID, err)
	}
	_, err = u.sensorRepository.GetSensorByID(ctx, sensorID)
	if err != nil {
		return fmt.Errorf("got invalid sensor id (%v): %w", userID, err)
	}

	return u.sensorOwnerRepository.SaveSensorOwner(ctx, domain.SensorOwner{
		UserID:   userID,
		SensorID: sensorID,
	})
}

func (u *User) GetUserSensors(ctx context.Context, userID int64) ([]domain.Sensor, error) {
	_, err := u.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("got invalid user id (%v): %w", userID, err)
	}

	soArr, err := u.sensorOwnerRepository.GetSensorsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot get sensors for required user %v: %w", userID, err)
	}

	result := make([]domain.Sensor, 0, len(soArr))
	for _, so := range soArr {
		s, err := u.sensorRepository.GetSensorByID(ctx, so.SensorID)
		if err != nil {
			return result, fmt.Errorf("error getting sensor from repository: %v, %w", so, err)
		}
		result = append(result, *s)
	}

	return result, nil
}
