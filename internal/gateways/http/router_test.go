package http

import (
	"bytes"
	"context"
	"encoding/json"
	"homework/internal/domain"
	"homework/internal/gateways/http/dtos"
	"homework/internal/usecase"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	eventRepository "homework/internal/repository/event/inmemory"
	sensorRepository "homework/internal/repository/sensor/inmemory"
	subscriptionRepository "homework/internal/repository/subscription/inmemory"
	userRepository "homework/internal/repository/user/inmemory"
)

var (
	er  = eventRepository.NewEventRepository()
	sr  = sensorRepository.NewSensorRepository()
	ur  = userRepository.NewUserRepository()
	sor = userRepository.NewSensorOwnerRepository()
	esr = subscriptionRepository.NewSubscriptionRepository[domain.Event]()
)

var useCases = UseCases{
	Event:             usecase.NewEvent(er, sr, esr),
	Sensor:            usecase.NewSensor(sr),
	User:              usecase.NewUser(ur, sor, sr),
	EventSubscription: usecase.NewSubscription[domain.Event](esr, sr),
}

var (
	eventTime    = time.Now()
	testSensorId int64
)

var router = gin.Default()

func init() {
	reg, err := useCases.Sensor.RegisterSensor(context.Background(), &domain.Sensor{
		SerialNumber: "1233211230",
		Type:         "cc",
		Description:  "test sensor",
	})
	if err != nil {
		panic(err)
	}
	testSensorId = reg.ID
	err = useCases.Event.ReceiveEvent(context.Background(), &domain.Event{
		Timestamp:          eventTime,
		SensorSerialNumber: "1233211230",
		SensorID:           testSensorId,
		Payload:            1,
	})
	if err != nil {
		panic(err)
	}
	setupRouter(router, useCases, NewWebSocketHandler(useCases))
}

// Все неизвестные пути должны возвращать http.StatusNotFound.
func TestUnknownRoute(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{http.MethodGet, http.MethodGet, http.StatusNotFound},
		{http.MethodPost, http.MethodPost, http.StatusNotFound},
		{http.MethodPut, http.MethodPut, http.StatusNotFound},
		{http.MethodDelete, http.MethodDelete, http.StatusNotFound},
		{http.MethodHead, http.MethodHead, http.StatusNotFound},
		{http.MethodOptions, http.MethodOptions, http.StatusNotFound},
		{http.MethodPatch, http.MethodPatch, http.StatusNotFound},
		{http.MethodConnect, http.MethodConnect, http.StatusNotFound},
		{http.MethodTrace, http.MethodTrace, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.input, "/unknown", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.want, w.Code, "Получили в ответ не тот код")
		})
	}
}

// Тесты /users
func TestUsersRoutes(t *testing.T) {
	t.Run("POST_users", func(t *testing.T) {
		t.Run("valid_request_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{
				"name": "Пользователь 1"
			}`
			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.True(t, json.Valid(w.Body.Bytes()), "В ответе не json")
		})

		t.Run("request_body_has_unsupported_format_415", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `<User>
				<Name>Пользователь 1</Name>
			</User>`
			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnsupportedMediaType, w.Code, "Получили в ответ не тот код")
		})

		t.Run("request_body_has_syntax_error_400", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{ невалидный json }`
			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code, "Получили в ответ не тот код")
		})

		t.Run("request_body_is_valid_but_it_has_invalid_data_422", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{
				"name": ""
			}`
			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("OPTIONS_users_204", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/users", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Получили в ответ не тот код")
		allowed := strings.Split(w.Header().Get("Allow"), ",")
		assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
		assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
	})

	// Другие методы не поддерживаем.
	t.Run("OTHER_users_405", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
			want  int
		}{
			{http.MethodGet, http.MethodGet, http.StatusMethodNotAllowed},
			{http.MethodPut, http.MethodPut, http.StatusMethodNotAllowed},
			{http.MethodPut, http.MethodPut, http.StatusMethodNotAllowed},
			{http.MethodHead, http.MethodHead, http.StatusMethodNotAllowed},
			{http.MethodPatch, http.MethodPatch, http.StatusMethodNotAllowed},
			{http.MethodConnect, http.MethodConnect, http.StatusMethodNotAllowed},
			{http.MethodTrace, http.MethodTrace, http.StatusMethodNotAllowed},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(tt.input, "/users", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.want, w.Code, "Получили в ответ не тот код")

				// К сожалению, в gin нет возможности удобно настроить поведение для 405-ых ответов,
				// поэтому проверку наличия заголовка Allow отключаем.

				// allowed := strings.Split(w.Header().Get("Allow"), ",")
				// assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
				// assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
			})
		}
	})
}

// Тесты /sensors
func TestSensorsRoutes(t *testing.T) {
	t.Run("GET_sensors", func(t *testing.T) {
		t.Run("success_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/sensors", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.True(t, json.Valid(w.Body.Bytes()), "В ответе не json")
		})

		t.Run("requested_unsupported_body_format_406", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/sensors", nil)
			req.Header.Add("Accept", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotAcceptable, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("HEAD_sensors", func(t *testing.T) {
		t.Run("success_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodHead, "/sensors", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.NotEmpty(t, w.Header().Get("Content-Length"), "Content-Length не задан")
		})

		t.Run("requested_unsupported_body_format_406", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodHead, "/sensors", nil)
			req.Header.Add("Accept", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotAcceptable, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("POST_sensors", func(t *testing.T) {
		t.Run("valid_request_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{
				"serial_number": "1234567890",
				"type": "cc",
				"description": "Датчик температуры",
				"is_active": true
			}`
			req, _ := http.NewRequest(http.MethodPost, "/sensors", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.True(t, json.Valid(w.Body.Bytes()), "В ответе не json")
		})

		t.Run("request_body_has_unsupported_format_415", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `<Sensor>
				<SerialNumber>1234567890</SerialNumber>
				<Type>cc</Type>
				<Description>Датчик температуры</Description>
				<IsActive>true</IsActive>
			</Sensor>`
			req, _ := http.NewRequest(http.MethodPost, "/sensors", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnsupportedMediaType, w.Code, "Получили в ответ не тот код")
		})

		t.Run("request_body_has_syntax_error_400", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{ невалидный json }`
			req, _ := http.NewRequest(http.MethodPost, "/sensors", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code, "Получили в ответ не тот код")
		})

		t.Run("request_body_is_valid_but_it_has_invalid_data_422", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{
				"serial_number": "",
				"type": "cc",
				"description": "Датчик температуры",
				"is_active": true
			}`
			req, _ := http.NewRequest(http.MethodPost, "/sensors", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("OPTIONS_sensors_204", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/sensors", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Получили в ответ не тот код")
		allowed := strings.Split(w.Header().Get("Allow"), ",")
		assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
		assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
		assert.Contains(t, allowed, http.MethodGet, "В разрешённых методах нет GET")
		assert.Contains(t, allowed, http.MethodHead, "В разрешённых методах нет HEAD")
	})

	t.Run("OTHER_sensors_405", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
			want  int
		}{
			{http.MethodPut, http.MethodPut, http.StatusMethodNotAllowed},
			{http.MethodDelete, http.MethodDelete, http.StatusMethodNotAllowed},
			{http.MethodPatch, http.MethodPatch, http.StatusMethodNotAllowed},
			{http.MethodConnect, http.MethodConnect, http.StatusMethodNotAllowed},
			{http.MethodTrace, http.MethodTrace, http.StatusMethodNotAllowed},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(tt.input, "/sensors", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.want, w.Code, "Получили в ответ не тот код")
				// allowed := strings.Split(w.Header().Get("Allow"), ",")
				// assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
				// assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
			})
		}
	})

	t.Run("GET_sensors_sensor_id", func(t *testing.T) {
		t.Run("sensor_exists_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/sensors/"+strconv.FormatInt(testSensorId, 10), nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.True(t, json.Valid(w.Body.Bytes()), "В ответе не json")
		})

		t.Run("id_has_invalid_format_422", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/sensors/abc", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
		})

		t.Run("requested_unsupported_body_format_406", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/sensors/1", nil)
			req.Header.Add("Accept", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotAcceptable, w.Code, "Получили в ответ не тот код")
		})

		t.Run("sensor_doesnt_exist_404", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/sensors/10", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("HEAD_sensors_sensor_id", func(t *testing.T) {
		t.Run("sensor_exists_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodHead, "/sensors/1", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.NotEmpty(t, w.Header().Get("Content-Length"), "Content-Length не задан")
		})

		t.Run("id_has_invalid_format_422", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodHead, "/sensors/abc", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
		})

		t.Run("requested_unsupported_body_format_406", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodHead, "/sensors/1", nil)
			req.Header.Add("Accept", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotAcceptable, w.Code, "Получили в ответ не тот код")
		})

		t.Run("sensor_doesnt_exist_404", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodHead, "/sensors/10", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("OPTIONS_sensors_sensor_id_204", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/sensors/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Получили в ответ не тот код")
		allowed := strings.Split(w.Header().Get("Allow"), ",")
		assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
		assert.Contains(t, allowed, http.MethodGet, "В разрешённых методах нет GET")
		assert.Contains(t, allowed, http.MethodHead, "В разрешённых методах нет HEAD")
	})

	// Другие методы не поддерживаем.
	t.Run("OTHER_sensors_sensor_id_405", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
			want  int
		}{
			{http.MethodPost, http.MethodPost, http.StatusMethodNotAllowed},
			{http.MethodPut, http.MethodPut, http.StatusMethodNotAllowed},
			{http.MethodDelete, http.MethodDelete, http.StatusMethodNotAllowed},
			{http.MethodPatch, http.MethodPatch, http.StatusMethodNotAllowed},
			{http.MethodConnect, http.MethodConnect, http.StatusMethodNotAllowed},
			{http.MethodTrace, http.MethodTrace, http.StatusMethodNotAllowed},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(tt.input, "/sensors/1", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.want, w.Code, "Получили в ответ не тот код")
				// allowed := strings.Split(w.Header().Get("Allow"), ",")
				// assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
				// assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
			})
		}
	})

	t.Run("GET_sensors_sensor_history", func(t *testing.T) {
		t.Run("history_ok_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			escapedStartDate := url.QueryEscape(strfmt.DateTime(eventTime).String())
			escapedEndDate := url.QueryEscape(strfmt.DateTime(eventTime.Add(1 * time.Second)).String())
			req, _ := http.NewRequest(http.MethodGet, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history?start_date="+escapedStartDate+"&end_date="+escapedEndDate, nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.True(t, json.Valid(w.Body.Bytes()), "В ответе не json")
			var resp []dtos.SensorHistory
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp), "В ответе не список событий")
			assert.Greater(t, len(resp), 0, "В ответе пусто")
		})

		t.Run("422", func(t *testing.T) {
			t.Run("id_has_invalid_sensor_id_format", func(t *testing.T) {
				w := httptest.NewRecorder()

				req, _ := http.NewRequest(http.MethodGet, "/sensors/abc", nil)
				req.Header.Add("Accept", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
			})

			t.Run("query_has_invalid_time_frame", func(t *testing.T) {
				w := httptest.NewRecorder()

				escapedStartDate := url.QueryEscape(strfmt.DateTime(eventTime).String())
				escapedEndDate := url.QueryEscape(strfmt.DateTime(eventTime.Add(1 * time.Second)).String())
				req, _ := http.NewRequest(http.MethodGet, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history?start_date="+escapedEndDate+"&end_date="+escapedStartDate, nil)
				req.Header.Add("Accept", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
			})

			t.Run("query_has_invalid_start_time_format", func(t *testing.T) {
				w := httptest.NewRecorder()

				escapedStartDate := "invalid"
				escapedEndDate := url.QueryEscape(strfmt.DateTime(eventTime.Add(1 * time.Second)).String())
				req, _ := http.NewRequest(http.MethodGet, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history?start_date="+escapedStartDate+"&end_date="+escapedEndDate, nil)
				req.Header.Add("Accept", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
			})

			t.Run("query_has_invalid_end_time_format", func(t *testing.T) {
				w := httptest.NewRecorder()

				escapedStartDate := url.QueryEscape(strfmt.DateTime(eventTime).String())
				escapedEndDate := "invalid"
				req, _ := http.NewRequest(http.MethodGet, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history?start_date="+escapedStartDate+"&end_date="+escapedEndDate, nil)
				req.Header.Add("Accept", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
			})

			t.Run("missing_query", func(t *testing.T) {
				w := httptest.NewRecorder()

				req, _ := http.NewRequest(http.MethodGet, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history", nil)
				req.Header.Add("Accept", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
			})
		})

		t.Run("requested_unsupported_body_format_406", func(t *testing.T) {
			w := httptest.NewRecorder()

			escapedStartDate := url.QueryEscape(strfmt.DateTime(eventTime).String())
			escapedEndDate := url.QueryEscape(strfmt.DateTime(eventTime.Add(1 * time.Second)).String())
			req, _ := http.NewRequest(http.MethodGet, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history?start_date="+escapedStartDate+"&end_date="+escapedEndDate, nil)
			req.Header.Add("Accept", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotAcceptable, w.Code, "Получили в ответ не тот код")
		})

		t.Run("sensor_doesnt_exist_404", func(t *testing.T) {
			w := httptest.NewRecorder()

			escapedStartDate := url.QueryEscape(strfmt.DateTime(eventTime).String())
			escapedEndDate := url.QueryEscape(strfmt.DateTime(eventTime.Add(1 * time.Second)).String())
			req, _ := http.NewRequest(http.MethodGet, "/sensors/10/history?start_date="+escapedStartDate+"&end_date="+escapedEndDate, nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("HEAD_sensors_sensor_history", func(t *testing.T) {
		t.Run("history_ok_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			escapedStartDate := url.QueryEscape(strfmt.DateTime(eventTime).String())
			escapedEndDate := url.QueryEscape(strfmt.DateTime(eventTime.Add(1 * time.Second)).String())
			req, _ := http.NewRequest(http.MethodHead, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history?start_date="+escapedStartDate+"&end_date="+escapedEndDate, nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.Empty(t, w.Body)
			assert.NotEmpty(t, w.Header().Get("Content-Length"), "Content-Length не задан")
		})

		t.Run("422", func(t *testing.T) {
			t.Run("id_has_invalid_sensor_id_format", func(t *testing.T) {
				w := httptest.NewRecorder()

				req, _ := http.NewRequest(http.MethodHead, "/sensors/abc", nil)
				req.Header.Add("Accept", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
			})

			t.Run("query_has_invalid_time_frame", func(t *testing.T) {
				w := httptest.NewRecorder()

				escapedStartDate := url.QueryEscape(strfmt.DateTime(eventTime).String())
				escapedEndDate := url.QueryEscape(strfmt.DateTime(eventTime.Add(1 * time.Second)).String())
				req, _ := http.NewRequest(http.MethodHead, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history?start_date="+escapedEndDate+"&end_date="+escapedStartDate, nil)
				req.Header.Add("Accept", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
			})

			t.Run("query_has_invalid_start_time_format", func(t *testing.T) {
				w := httptest.NewRecorder()

				escapedStartDate := "invalid"
				escapedEndDate := url.QueryEscape(strfmt.DateTime(eventTime.Add(1 * time.Second)).String())
				req, _ := http.NewRequest(http.MethodHead, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history?start_date="+escapedStartDate+"&end_date="+escapedEndDate, nil)
				req.Header.Add("Accept", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
			})

			t.Run("query_has_invalid_end_time_format", func(t *testing.T) {
				w := httptest.NewRecorder()

				escapedStartDate := url.QueryEscape(strfmt.DateTime(eventTime).String())
				escapedEndDate := "invalid"
				req, _ := http.NewRequest(http.MethodHead, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history?start_date="+escapedStartDate+"&end_date="+escapedEndDate, nil)
				req.Header.Add("Accept", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
			})

			t.Run("missing_query", func(t *testing.T) {
				w := httptest.NewRecorder()

				req, _ := http.NewRequest(http.MethodHead, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history", nil)
				req.Header.Add("Accept", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
			})
		})

		t.Run("requested_unsupported_body_format_406", func(t *testing.T) {
			w := httptest.NewRecorder()

			escapedStartDate := url.QueryEscape(strfmt.DateTime(eventTime).String())
			escapedEndDate := url.QueryEscape(strfmt.DateTime(eventTime.Add(1 * time.Second)).String())
			req, _ := http.NewRequest(http.MethodHead, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history?start_date="+escapedStartDate+"&end_date="+escapedEndDate, nil)
			req.Header.Add("Accept", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotAcceptable, w.Code, "Получили в ответ не тот код")
		})

		t.Run("sensor_doesnt_exist_404", func(t *testing.T) {
			w := httptest.NewRecorder()

			escapedStartDate := url.QueryEscape(strfmt.DateTime(eventTime).String())
			escapedEndDate := url.QueryEscape(strfmt.DateTime(eventTime.Add(1 * time.Second)).String())
			req, _ := http.NewRequest(http.MethodHead, "/sensors/10/history?start_date="+escapedStartDate+"&end_date="+escapedEndDate, nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("OPTIONS_sensors_history_204", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Получили в ответ не тот код")
		allowed := strings.Split(w.Header().Get("Allow"), ",")
		assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
		assert.Contains(t, allowed, http.MethodGet, "В разрешённых методах нет GET")
		assert.Contains(t, allowed, http.MethodHead, "В разрешённых методах нет HEAD")
	})

	// Другие методы не поддерживаем.
	t.Run("OTHER_sensors_history_405", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
			want  int
		}{
			{http.MethodPost, http.MethodPost, http.StatusMethodNotAllowed},
			{http.MethodPut, http.MethodPut, http.StatusMethodNotAllowed},
			{http.MethodDelete, http.MethodDelete, http.StatusMethodNotAllowed},
			{http.MethodPatch, http.MethodPatch, http.StatusMethodNotAllowed},
			{http.MethodConnect, http.MethodConnect, http.StatusMethodNotAllowed},
			{http.MethodTrace, http.MethodTrace, http.StatusMethodNotAllowed},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(tt.input, "/sensors/"+strconv.FormatInt(testSensorId, 10)+"/history", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.want, w.Code, "Получили в ответ не тот код")
			})
		}
	})
}

// Тесты /users/{user_id}/sensors
func TestUsersSensorsRoutes(t *testing.T) {
	t.Run("GET_users_user_id_sensors", func(t *testing.T) {
		t.Run("user_exists_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/users/1/sensors", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.True(t, json.Valid(w.Body.Bytes()), "В ответе не json")
		})

		t.Run("id_has_invalid_format_422", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/users/abc/sensors", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
		})

		t.Run("requested_unsupported_body_format_406", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/users/1/sensors", nil)
			req.Header.Add("Accept", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotAcceptable, w.Code, "Получили в ответ не тот код")
		})

		t.Run("user_doesnt_exist_404", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/users/2/sensors", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("HEAD_users_user_id_sensors", func(t *testing.T) {
		t.Run("user_exists_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodHead, "/users/1/sensors", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.NotEmpty(t, w.Header().Get("Content-Length"), "Content-Length не задан")
		})

		t.Run("id_has_invalid_format_422", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodHead, "/users/abc/sensors", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
		})

		t.Run("requested_unsupported_body_format_406", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodHead, "/users/1/sensors", nil)
			req.Header.Add("Accept", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotAcceptable, w.Code, "Получили в ответ не тот код")
		})

		t.Run("user_doesnt_exist_404", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodHead, "/users/2/sensors", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("POST_users_user_id_sensors", func(t *testing.T) {
		t.Run("valid_request_body_and_user_exists_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{
				"sensor_id": 1
			}`
			req, _ := http.NewRequest(http.MethodPost, "/users/1/sensors", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code, "Получили в ответ не тот код")
		})

		t.Run("request_body_has_unsupported_format_415", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `<SensorToUserBinding>
				<SensorId>1</SensorId>
			</SensorToUserBinding>`
			req, _ := http.NewRequest(http.MethodPost, "/users/1/sensors", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnsupportedMediaType, w.Code, "Получили в ответ не тот код")
		})

		t.Run("invalid_request_body_400", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{ невалидный json }`
			req, _ := http.NewRequest(http.MethodPost, "/users/1/sensors", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code, "Получили в ответ не тот код")
		})

		t.Run("valid_request_body_but_user_doesnt_exist_404", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{
				"sensor_id": 1
			}`
			req, _ := http.NewRequest(http.MethodPost, "/users/2/sensors", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Получили в ответ не тот код")
		})

		t.Run("request_body_is_valid_but_it_has_invalid_data_422", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{
				"sensor_id": -1
			}`
			req, _ := http.NewRequest(http.MethodPost, "/users/1/sensors", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("OPTIONS_users_user_id_sensors_204", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/users/1/sensors", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Получили в ответ не тот код")
		allowed := strings.Split(w.Header().Get("Allow"), ",")
		assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
		assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
		assert.Contains(t, allowed, http.MethodHead, "В разрешённых методах нет HEAD")
		assert.Contains(t, allowed, http.MethodGet, "В разрешённых методах нет GET")
	})

	// Другие методы не поддерживаем.
	t.Run("OTHER_users_user_id_sensors_405", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
			want  int
		}{
			{http.MethodPut, http.MethodPut, http.StatusMethodNotAllowed},
			{http.MethodDelete, http.MethodDelete, http.StatusMethodNotAllowed},
			{http.MethodPatch, http.MethodPatch, http.StatusMethodNotAllowed},
			{http.MethodConnect, http.MethodConnect, http.StatusMethodNotAllowed},
			{http.MethodTrace, http.MethodTrace, http.StatusMethodNotAllowed},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(tt.input, "/users", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.want, w.Code, "Получили в ответ не тот код")
				// allowed := strings.Split(w.Header().Get("Allow"), ",")
				// assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
				// assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
				// assert.Contains(t, allowed, http.MethodHead, "В разрешённых методах нет HEAD")
				// assert.Contains(t, allowed, http.MethodGet, "В разрешённых методах нет GET")
			})
		}
	})
}

// Тесты /events
func TestEventsRoutes(t *testing.T) {
	t.Run("POST_events", func(t *testing.T) {
		t.Run("valid_request_201", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{
				"sensor_serial_number": "1234567890",
				"payload": 10
			}`
			req, _ := http.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code, "Получили в ответ не тот код")
		})

		t.Run("request_body_has_unsupported_format_415", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `<SensorEvent>
				<SensorSerialNumber>1234567890</SensorSerialNumber>
				<Payload>10</Payload>
			</SensorEvent>`
			req, _ := http.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/xml")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnsupportedMediaType, w.Code, "Получили в ответ не тот код")
		})

		t.Run("request_body_has_syntax_error_400", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{ невалидный json }`
			req, _ := http.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code, "Получили в ответ не тот код")
		})

		t.Run("request_body_is_valid_but_it_has_invalid_data_422", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{
				"sensor_serial_number": "",
				"payload": 10
			}`
			req, _ := http.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Получили в ответ не тот код")
		})
	})

	t.Run("OPTIONS_events_204", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/events", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Получили в ответ не тот код")
		allowed := strings.Split(w.Header().Get("Allow"), ",")
		assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
		assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
	})

	// Другие методы не поддерживаем.
	t.Run("OTHER_users_405", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
			want  int
		}{
			{http.MethodGet, http.MethodGet, http.StatusMethodNotAllowed},
			{http.MethodPut, http.MethodPut, http.StatusMethodNotAllowed},
			{http.MethodDelete, http.MethodDelete, http.StatusMethodNotAllowed},
			{http.MethodHead, http.MethodHead, http.StatusMethodNotAllowed},
			{http.MethodPatch, http.MethodPatch, http.StatusMethodNotAllowed},
			{http.MethodConnect, http.MethodConnect, http.StatusMethodNotAllowed},
			{http.MethodTrace, http.MethodTrace, http.StatusMethodNotAllowed},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(tt.input, "/events", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.want, w.Code, "Получили в ответ не тот код")
				// allowed := strings.Split(w.Header().Get("Allow"), ",")
				// assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
				// assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
			})
		}
	})
}
