package usecase

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"regexp"
)

type Sensor struct {
	sensorRepository SensorRepository
}

func NewSensor(sr SensorRepository) *Sensor {
	return &Sensor{
		sensorRepository: sr,
	}
}

func (s *Sensor) validateSN(sn string) bool {
	re := regexp.MustCompile(`^[0-9]{10}$`)
	return re.MatchString(sn)
}

func (s *Sensor) validateSensType(t domain.SensorType) bool {
	return t == domain.SensorTypeADC ||
		t == domain.SensorTypeContactClosure
}

func (s *Sensor) RegisterSensor(ctx context.Context, sensor *domain.Sensor) (*domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return sensor, err
	}
	if sensor == nil {
		return sensor, errors.New("got nil sensor at RegisterSensor()")
	}
	if !s.validateSN(sensor.SerialNumber) {
		return sensor, ErrWrongSensorSerialNumber
	}
	if !s.validateSensType(sensor.Type) {
		return sensor, ErrWrongSensorType
	}

	if existing, err := s.sensorRepository.GetSensorBySerialNumber(ctx, sensor.SerialNumber); err == nil {
		existing.LastActivity = sensor.LastActivity
		existing.CurrentState = sensor.CurrentState
		existing.Description = sensor.Description
		existing.IsActive = sensor.IsActive
		return existing, nil
	} else if !errors.Is(err, ErrSensorNotFound) { // bad design :(, consider refactoring interface to invert that dependency
		return nil, err
	}

	if err := s.sensorRepository.SaveSensor(ctx, sensor); err != nil { // Modifies sensor assigning new id
		return sensor, err
	}
	return sensor, nil
}

func (s *Sensor) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sens, err := s.sensorRepository.GetSensors(ctx)
	if err != nil {
		return sens, fmt.Errorf("cannot get sensors from repository: %w", err)
	}
	return sens, err
}

func (s *Sensor) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sens, err := s.sensorRepository.GetSensorByID(ctx, id)
	if err != nil {
		return sens, fmt.Errorf("cannot get sensors from repository for id %v: %w", id, err)
	}

	return sens, nil
}
