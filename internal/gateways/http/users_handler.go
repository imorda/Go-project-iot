package http

import (
	"errors"
	"homework/internal/domain"
	"homework/internal/gateways/http/dtos"
	"homework/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func userGetImpl(sensor *domain.User) dtos.User {
	userDto := dtos.User{
		ID:   &sensor.ID,
		Name: &sensor.Name,
	}
	return userDto
}

func usersPostHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userDto := &dtos.UserToCreate{}
		if extractDto(ctx, userDto) == nil {
			userToCreate := domain.User{
				Name: *userDto.Name,
			}

			userCreated, err := uc.User.RegisterUser(ctx, &userToCreate)
			if err != nil {
				abortWithAPIError(ctx, http.StatusInternalServerError, err)
				return
			}

			ctx.AbortWithStatusJSON(http.StatusOK, userGetImpl(userCreated))
		}
	}
}

func userSensorsCommonHandler(ctx *gin.Context, uc UseCases) (*int64, []dtos.Sensor) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		abortWithAPIError(ctx, http.StatusUnprocessableEntity, err)
		return nil, nil
	}

	sensors, err := uc.User.GetUserSensors(ctx, userId)
	if errors.Is(err, usecase.ErrUserNotFound) {
		ctx.AbortWithStatus(http.StatusNotFound)
		return &userId, nil
	} else if err != nil {
		abortWithAPIError(ctx, http.StatusInternalServerError, err)
		return &userId, nil
	}

	sensorDtos := make([]dtos.Sensor, 0, len(sensors))
	for _, sens := range sensors {
		sensDto := sensorGetImpl(&sens)
		sensorDtos = append(sensorDtos, sensDto)
	}

	return &userId, sensorDtos
}

func userSensorsGetHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := isFormatSupported(ctx, JSONType); err != nil {
			abortWithAPIError(ctx, http.StatusNotAcceptable, err)
			return
		}

		_, sensorDtos := userSensorsCommonHandler(ctx, uc)
		if !ctx.IsAborted() {
			ctx.AbortWithStatusJSON(http.StatusOK, sensorDtos)
		}
	}
}

func userSensorsHeadHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := isFormatSupported(ctx, JSONType); err != nil {
			abortWithAPIError(ctx, http.StatusNotAcceptable, err)
			return
		}

		_, sensorDtos := userSensorsCommonHandler(ctx, uc)
		if !ctx.IsAborted() {
			headImpl(ctx, sensorDtos)
		}
	}
}

func userSensorPostHandler(uc UseCases) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, _ := userSensorsCommonHandler(ctx, uc)
		if ctx.IsAborted() {
			return
		}

		bindingDto := &dtos.SensorToUserBinding{}
		if extractDto(ctx, bindingDto) == nil {
			err := uc.User.AttachSensorToUser(ctx, *userId, *bindingDto.SensorID)
			if err != nil {
				abortWithAPIError(ctx, http.StatusInternalServerError, err)
				return
			}

			ctx.Status(http.StatusCreated)
		}
	}
}

func setupUsersHandler(r *gin.RouterGroup, uc UseCases) {
	r.POST("", usersPostHandler(uc))
	r.OPTIONS("", optionsHandler(http.MethodPost))

	r.GET("/:user_id/sensors", userSensorsGetHandler(uc))
	r.HEAD("/:user_id/sensors", userSensorsHeadHandler(uc))
	r.POST("/:user_id/sensors", userSensorPostHandler(uc))
	r.OPTIONS("/:user_id/sensors", optionsHandler(http.MethodGet, http.MethodHead, http.MethodPost))
}
