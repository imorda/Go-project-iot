package http

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/metrics"
	"homework/internal/usecase"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	host   string
	port   uint16
	router *gin.Engine
	wsh    *WebSocketHandler
}

type UseCases struct {
	Event             *usecase.Event
	Sensor            *usecase.Sensor
	User              *usecase.User
	EventSubscription *usecase.Subscription[domain.Event]
}

func NewServer(useCases UseCases, options ...func(*Server)) *Server {
	r := gin.Default()
	r.HandleMethodNotAllowed = true
	r.Use(metrics.RedMiddleware())
	apiGroup := r.Group("/api")
	wsh := NewWebSocketHandler(useCases)
	setupRouter(apiGroup, useCases, wsh)

	s := &Server{router: r, host: "localhost", port: 8080, wsh: wsh}
	for _, o := range options {
		o(s)
	}

	return s
}

func WithHost(host string) func(*Server) {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port uint16) func(*Server) {
	return func(s *Server) {
		s.port = port
	}
}

func (s *Server) Run(ctx context.Context, cancel context.CancelFunc) error {
	serverErr := make(chan error)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: s.router,
	}

	go func() {
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server crashed: %w", err)
	case <-ctx.Done():
		log.Printf("Gracefully shutting down the server...")
		composedErr := errors.Join(s.wsh.Shutdown(), srv.Shutdown(ctx))
		cancel()
		return composedErr
	}
}
