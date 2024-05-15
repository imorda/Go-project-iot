package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

type SensorRepository struct {
	idStorage map[int64]*domain.Sensor
	snStorage map[string]*domain.Sensor
	lastId    int64
	mu        sync.Mutex
}

func NewSensorRepository() *SensorRepository {
	return &SensorRepository{
		idStorage: make(map[int64]*domain.Sensor),
		snStorage: make(map[string]*domain.Sensor),
		lastId:    0,
		mu:        sync.Mutex{},
	}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if sensor == nil {
		return errors.New("got nil sensor at SaveSensor()")
	}
	if sensor.RegisteredAt.IsZero() {
		sensor.RegisteredAt = time.Now()

		r.mu.Lock()
		r.lastId++
		id := r.lastId
		sensor.ID = id

		r.idStorage[id] = sensor
		r.snStorage[sensor.SerialNumber] = sensor
		r.mu.Unlock()
	}

	return nil
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	res := make([]domain.Sensor, 0, len(r.idStorage))

	for _, v := range r.idStorage {
		res = append(res, *v)
	}
	r.mu.Unlock()

	return res, nil
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	val, exists := r.idStorage[id]
	r.mu.Unlock()
	if !exists {
		return nil, usecase.ErrSensorNotFound
	}

	return val, nil
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	val, exists := r.snStorage[sn]
	r.mu.Unlock()
	if !exists {
		return nil, usecase.ErrSensorNotFound
	}

	return val, nil
}
