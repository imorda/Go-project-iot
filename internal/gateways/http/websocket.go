package http

import (
	"context"
	"errors"
	"homework/internal/metrics"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"nhooyr.io/websocket"
)

const (
	WriteTimeout = 5 * time.Second
	BatchPeriod  = 500 * time.Millisecond
)

type WebSocketHandler struct {
	useCases UseCases
	conns    map[*websocket.Conn]struct{}
	mu       sync.Mutex
}

func NewWebSocketHandler(useCases UseCases) *WebSocketHandler {
	return &WebSocketHandler{
		useCases: useCases,
		conns:    make(map[*websocket.Conn]struct{}),
		mu:       sync.Mutex{},
	}
}

func (h *WebSocketHandler) HandleSubscription(c *gin.Context, ch <-chan []byte) error {
	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	if err != nil {
		return err
	}

	routineErr := func() error {
		connCtx := conn.CloseRead(c)
		h.mu.Lock()
		h.conns[conn] = struct{}{}
		metrics.AddCounter("ws_connections", 1)
		h.mu.Unlock()

		for {
			select {
			case msg := <-ch:
				ctxTimed, cancel := context.WithTimeout(connCtx, WriteTimeout)
				err := conn.Write(ctxTimed, websocket.MessageText, msg)
				cancel()
				if err != nil {
					return err
				}
			case <-connCtx.Done():
				return connCtx.Err()
			case <-c.Done():
				return c.Err()
			}
		}
	}()

	h.mu.Lock()
	var closeErr error
	if _, ok := h.conns[conn]; ok {
		closeErr = conn.CloseNow()
		delete(h.conns, conn)
		metrics.AddCounter("ws_connections", -1)
	}
	h.mu.Unlock()

	return errors.Join(routineErr, closeErr)
}

func (h *WebSocketHandler) Shutdown() error {
	log.Printf("Shutting down ws...")
	h.mu.Lock()
	defer h.mu.Unlock()

	var errs []error

	for conn := range h.conns {
		err := conn.Close(websocket.StatusNormalClosure, "server shutting down")
		metrics.AddCounter("ws_connections", -1)
		if err != nil {
			errs = append(errs, err)
		}
	}
	h.conns = map[*websocket.Conn]struct{}{}
	return errors.Join(errs...)
}
