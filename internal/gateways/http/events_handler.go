package http

import (
	"homework/internal/domain"
	"homework/internal/gateways/http/dtos"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
	r.POST("", eventsPostHandler(uc))
	r.OPTIONS("", optionsHandler(http.MethodPost))
}
