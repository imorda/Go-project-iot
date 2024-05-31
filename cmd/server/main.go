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

	"github.com/jackc/pgx/v5/pgxpool"

	httpGateway "homework/internal/gateways/http"
	metrics "homework/internal/metrics"
	eventRepository "homework/internal/repository/event/postgres"
	sensorRepository "homework/internal/repository/sensor/postgres"
	subscriptionRepository "homework/internal/repository/subscription/inmemory"
	userRepository "homework/internal/repository/user/postgres"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("can't parse pgxpool config")
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("can't create new pool")
	}
	defer pool.Close()

	er := eventRepository.NewEventRepository(pool)
	sr := sensorRepository.NewSensorRepository(pool)
	ur := userRepository.NewUserRepository(pool)
	sor := userRepository.NewSensorOwnerRepository(pool)
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

	metricsPort := 8008
	if port, ok := os.LookupEnv("METRICS_PORT"); ok {
		metricsPort, err = strconv.Atoi(port)
		if err != nil {
			log.Fatalf("invalid METRICS_PORT is set: %v", err)
		}
	}
	metrics.InitMetricsServer(metricsPort)

	if err := r.Run(ctx, cancel); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error during server shutdown: %v", err)
	}
}
