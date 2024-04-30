package main

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	httpGateway "homework/internal/gateways/http"
	eventRepository "homework/internal/repository/event/inmemory"
	sensorRepository "homework/internal/repository/sensor/inmemory"
	subscriptionRepository "homework/internal/repository/subscription/inmemory"
	userRepository "homework/internal/repository/user/inmemory"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	er := eventRepository.NewEventRepository()
	sr := sensorRepository.NewSensorRepository()
	ur := userRepository.NewUserRepository()
	sor := userRepository.NewSensorOwnerRepository()
	esr := subscriptionRepository.NewSubscriptionRepository[domain.Event]()

	useCases := httpGateway.UseCases{
		Event:             usecase.NewEvent(er, sr, esr),
		Sensor:            usecase.NewSensor(sr),
		User:              usecase.NewUser(ur, sor, sr),
		EventSubscription: usecase.NewSubscription[domain.Event](esr, sr),
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

	if err := r.Run(ctx, cancel); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error during server shutdown: %v", err)
	}
}
