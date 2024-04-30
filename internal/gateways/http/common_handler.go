package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/gateways/http/dtos"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/strfmt"
)

type ContentType = string

const ( // Supported Content-Type
	JSONType ContentType = "application/json"
	TextType ContentType = "plain/text"
)

var ( // Errors
	UnsupportedContentType = errors.New("unsupported Content-Type")
	UnsupportedAcceptType  = errors.New("unsupported Accept type")
)

func extractDto(ctx *gin.Context, dto Validator) error {
	if err := requireContentType(ctx, JSONType); err != nil {
		abortWithAPIError(ctx, http.StatusUnsupportedMediaType, err)
		return err
	}

	if err := ctx.BindJSON(dto); err != nil {
		return err // Return 400
	}

	if err := dto.Validate(nil); err != nil {
		abortWithAPIError(ctx, http.StatusUnprocessableEntity, err)
		return err
	}
	return nil
}

func isFormatSupported(ctx *gin.Context, contentType ...ContentType) error {
	expType := ctx.GetHeader("Accept")
	expCategories := strings.Split(expType, "/")
	if len(expCategories) != 2 {
		return nil
	}
	if expCategories[0] == "*" && expCategories[1] == "*" {
		return nil
	}

	for _, ct := range contentType {
		if expCategories[1] == "*" && strings.HasPrefix(ct, expCategories[0]) {
			return nil
		}
		if expType == ct {
			return nil
		}
	}
	return fmt.Errorf("%w requested: %v", UnsupportedAcceptType, expType)
}

func formatAPIError(err error) *dtos.Error {
	errReason := err.Error()
	dto := dtos.Error{Reason: &errReason}
	return &dto
}

func abortWithAPIError(ctx *gin.Context, code int, err error) {
	ctx.AbortWithStatusJSON(code, formatAPIError(err))
}

type Validator = interface {
	Validate(formats strfmt.Registry) error
}

func optionsHandler(methods ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Allow", strings.Join(append(methods, http.MethodOptions), ","))
		ctx.Status(http.StatusNoContent)
	}
}

func requireContentType(ctx *gin.Context, contentType ...ContentType) error {
	realType := ctx.GetHeader("Content-Type")
	for _, ct := range contentType {
		if realType == ct {
			return nil
		}
	}
	return fmt.Errorf("got %w: %v", UnsupportedContentType, realType)
}

func headImpl(ctx *gin.Context, jsonObj any) {
	if sensorBytes, err := json.Marshal(jsonObj); err == nil {
		ctx.Header("Content-Length", strconv.Itoa(len(sensorBytes)))
		ctx.Status(http.StatusOK)
	}
}

func eventChannelAdapter(ec <-chan domain.Event) <-chan []byte {
	out := make(chan []byte)
	go func() {
		defer close(out)
		for i := range ec {
			encoded, err := json.Marshal(&i)
			if err != nil {
				log.Printf("unable to encode event %v: %v", i.SensorID, err)
				return
			}
			out <- encoded
		}
	}()
	return out
}

func channelBatcher[T any](ec <-chan T, period time.Duration) <-chan T {
	toSend := time.NewTicker(period)
	out := make(chan T)
	go func() {
		defer toSend.Stop()
		defer close(out)

		var buffer []T
		flushBuffer := func() {
			for _, i := range buffer {
				out <- i
			}
			buffer = nil
		}

		for {
			select {
			case readVal, ok := <-ec:
				if !ok {
					flushBuffer()
					break
				}
				buffer = append(buffer, readVal)
			case <-toSend.C:
				flushBuffer()
			}
		}
	}()
	return out
}
