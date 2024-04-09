package main

import (
	"errors"
	"log"
	"net/http"

	httpGateway "homework/internal/gateways/http"
	eventRepository "homework/internal/repository/event/inmemory"
	sensorRepository "homework/internal/repository/sensor/inmemory"
	userRepository "homework/internal/repository/user/inmemory"
	"homework/internal/usecase"
)

func main() {
	er := eventRepository.NewEventRepository()
	sr := sensorRepository.NewSensorRepository()
	ur := userRepository.NewUserRepository()
	sor := userRepository.NewSensorOwnerRepository()

	useCases := httpGateway.UseCases{
		Event:  usecase.NewEvent(er, sr),
		Sensor: usecase.NewSensor(sr),
		User:   usecase.NewUser(ur, sor, sr),
	}

	// TODO реализовать веб-сервис

	r := httpGateway.NewServer(useCases)
	if err := r.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error during server shutdown: %v", err)
	}
}
