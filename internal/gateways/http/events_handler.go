package http

import (
	"github.com/gin-gonic/gin"
	"homework/internal/domain"
	"homework/internal/gateways/http/dtos"
	"net/http"
	"time"
)

func eventsPostHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		eventDto := &dtos.SensorEvent{}
		if extractDto(ctx, eventDto) == nil {
			event := domain.Event{
				Timestamp:          time.Now(),
				SensorSerialNumber: *eventDto.SensorSerialNumber,
				Payload:            *eventDto.Payload,
			}

			if err := uc.Event.ReceiveEvent(ctx, &event); err != nil {
				abortWithAPIError(ctx, http.StatusInternalServerError, err)
				return
			}

			ctx.Status(http.StatusCreated)
		}
	}
}

func setupEventsHandler(r *gin.RouterGroup, uc UseCases) {
	r.POST("/events", eventsPostHandler(uc))
	r.OPTIONS("/events", optionsHandler(http.MethodPost))
}
