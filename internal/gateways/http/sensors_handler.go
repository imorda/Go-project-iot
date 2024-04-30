package http

import (
	"errors"
	"homework/internal/domain"
	"homework/internal/gateways/http/dtos"
	"homework/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/strfmt"
)

func sensorGetImpl(sensor *domain.Sensor) dtos.Sensor {
	lastActivity := strfmt.DateTime(sensor.LastActivity)
	registeredAt := strfmt.DateTime(sensor.RegisteredAt)
	sensorType := string(sensor.Type)
	sensorDto := dtos.Sensor{
		CurrentState: &sensor.CurrentState,
		Description:  &sensor.Description,
		ID:           &sensor.ID,
		IsActive:     &sensor.IsActive,
		LastActivity: &lastActivity,
		RegisteredAt: &registeredAt,
		SerialNumber: &sensor.SerialNumber,
		Type:         &sensorType,
	}
	return sensorDto
}

func sensorsGetImpl(ctx *gin.Context, uc UseCases) []dtos.Sensor {
	if err := isFormatSupported(ctx, JSONType); err != nil {
		abortWithAPIError(ctx, http.StatusNotAcceptable, err)
		return nil
	}

	sensors, err := uc.Sensor.GetSensors(ctx)
	if err != nil {
		abortWithAPIError(ctx, http.StatusInternalServerError, err)
		return nil
	}

	sensorDtos := make([]dtos.Sensor, 0, len(sensors))
	for _, sens := range sensors {
		sensDto := sensorGetImpl(&sens)
		sensorDtos = append(sensorDtos, sensDto)
	}
	return sensorDtos
}

func sensorsGetHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensorDtos := sensorsGetImpl(ctx, uc)
		if !ctx.IsAborted() {
			ctx.AbortWithStatusJSON(http.StatusOK, sensorDtos)
		}
	}
}

func sensorsHeadHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensorDtos := sensorsGetImpl(ctx, uc)
		if !ctx.IsAborted() {
			headImpl(ctx, sensorDtos)
		}
	}
}

func sensorsPostHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensorDto := &dtos.SensorToCreate{}
		if extractDto(ctx, sensorDto) == nil {
			sensor := domain.Sensor{
				SerialNumber: *sensorDto.SerialNumber,
				Type:         domain.SensorType(*sensorDto.Type),
				Description:  *sensorDto.Description,
				IsActive:     *sensorDto.IsActive,
			}

			sens, err := uc.Sensor.RegisterSensor(ctx, &sensor)
			if err != nil {
				abortWithAPIError(ctx, http.StatusInternalServerError, err)
				return
			}

			ctx.AbortWithStatusJSON(http.StatusOK, sensorGetImpl(sens))
		}
	}
}

func sensorByIdCommonHandler(ctx *gin.Context, uc UseCases) *domain.Sensor {
	if err := isFormatSupported(ctx, JSONType); err != nil {
		abortWithAPIError(ctx, http.StatusNotAcceptable, err)
		return nil
	}

	sensorId, err := strconv.ParseInt(ctx.Param("sensor_id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnprocessableEntity)
		return nil
	}

	sensor, err := uc.Sensor.GetSensorByID(ctx, sensorId)
	if errors.Is(err, usecase.ErrSensorNotFound) {
		ctx.AbortWithStatus(http.StatusNotFound)
		return nil
	} else if err != nil {
		abortWithAPIError(ctx, http.StatusInternalServerError, err)
		return nil
	}

	return sensor
}

func sensorByIdGetHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensor := sensorByIdCommonHandler(ctx, uc)
		if !ctx.IsAborted() {
			ctx.AbortWithStatusJSON(http.StatusOK, sensorGetImpl(sensor))
		}
	}
}

func sensorByIdHeadHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensor := sensorByIdCommonHandler(ctx, uc)
		if !ctx.IsAborted() {
			headImpl(ctx, sensorGetImpl(sensor))
		}
	}
}

func setupSensorsHandler(r *gin.RouterGroup, uc UseCases) {
	r.GET("", sensorsGetHandler(uc))
	r.HEAD("", sensorsHeadHandler(uc))
	r.POST("", sensorsPostHandler(uc))
	r.OPTIONS("", optionsHandler(http.MethodGet, http.MethodHead, http.MethodPost))

	r.GET("/:sensor_id", sensorByIdGetHandler(uc))
	r.HEAD("/:sensor_id", sensorByIdHeadHandler(uc))
	r.OPTIONS("/:sensor_id", optionsHandler(http.MethodGet, http.MethodHead))
}
