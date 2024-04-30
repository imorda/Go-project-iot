package main

import (
	"errors"
	httpGateway "homework/internal/gateways/http"
	eventRepository "homework/internal/repository/event/inmemory"
	sensorRepository "homework/internal/repository/sensor/inmemory"
	userRepository "homework/internal/repository/user/inmemory"
	"homework/internal/usecase"
	"log"
	"net/http"
	"os"
	"strconv"
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

	r := httpGateway.NewServer(useCases)
	if host, ok := os.LookupEnv("HTTP_HOST"); ok {
		httpGateway.WithHost(host)(r)
	}
	if port, ok := os.LookupEnv("HTTP_PORT"); ok {
		iPort, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			log.Fatalf("invalid HTTP_PORT is set: %v", err)
		}
		httpGateway.WithPort(uint16(iPort))(r)
	}

	if err := r.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error during server shutdown: %v", err)
	}
}
