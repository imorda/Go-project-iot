package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	eventRepository "homework/internal/repository/event/inmemory"
	sensorRepository "homework/internal/repository/sensor/inmemory"
	userRepository "homework/internal/repository/user/inmemory"
	"homework/internal/usecase"
)

var (
	er  = eventRepository.NewEventRepository()
	sr  = sensorRepository.NewSensorRepository()
	ur  = userRepository.NewUserRepository()
	sor = userRepository.NewSensorOwnerRepository()
)

var useCases = UseCases{
	Event:  usecase.NewEvent(er, sr),
	Sensor: usecase.NewSensor(sr),
	User:   usecase.NewUser(ur, sor, sr),
}

var router = gin.Default()

func init() {
	setupRouter(router, useCases)
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
			req, _ := http.NewRequest(http.MethodGet, "/unknown", nil)
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
				<Id>1</Id>
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
				"id": -1,
				"name": "Пользователь -1"
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
				allowed := strings.Split(w.Header().Get("Allow"), ",")
				assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
				assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
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
				"serial_number": 1234567890,
				"type": "cc",
				"description": "Датчик температуры",
				"is_active": true,
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
				<Id>1</Id>
				<SerialNumber>1234567890</SerialNumber>
				<Type>cc</Type>
				<CurrentState>1</CurrentState>
				<Description>Датчик температуры</Description>
				<IsActive>true</IsActive>
				<RegisteredAt>2018-01-01T00:00:00Z</RegisteredAt>
				<LastActivity>2018-01-01T00:00:00Z</LastActivity>
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
				"id": -1,
				"serial_number": 1234567890,
				"type": "cc",
				"current_state": 1,
				"description": "Датчик температуры",
				"is_active": true,
				"registered_at": "2018-01-01T00:00:00Z",
				"last_activity": "2018-01-01T00:00:00Z"
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

	t.Run("GET_sensors_sensor_id", func(t *testing.T) {
		t.Run("sensor_exists_200", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/sensors/1", nil)
			req.Header.Add("Accept", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Получили в ответ не тот код")
			assert.True(t, json.Valid(w.Body.Bytes()), "В ответе не json")
		})

		t.Run("id_has_invalid_format_422", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, _ := http.NewRequest(http.MethodGet, "/sensors/1", nil)
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

			req, _ := http.NewRequest(http.MethodGet, "/sensors/2", nil)
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

			req, _ := http.NewRequest(http.MethodHead, "/sensors/1", nil)
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

			req, _ := http.NewRequest(http.MethodHead, "/sensors/2", nil)
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
				req, _ := http.NewRequest(tt.input, "/sensors", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.want, w.Code, "Получили в ответ не тот код")
				allowed := strings.Split(w.Header().Get("Allow"), ",")
				assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
				assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
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
				"sensor_id": 1,
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
				"sensor_id": 1,
			}`
			req, _ := http.NewRequest(http.MethodPost, "/users/2/sensors", bytes.NewReader([]byte(body)))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Получили в ответ не тот код")
		})

		t.Run("request_body_is_valid_but_it_has_invalid_data_422", func(t *testing.T) {
			w := httptest.NewRecorder()

			body := `{
				"sensor_id": -1,
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
				allowed := strings.Split(w.Header().Get("Allow"), ",")
				assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
				assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
				assert.Contains(t, allowed, http.MethodHead, "В разрешённых методах нет HEAD")
				assert.Contains(t, allowed, http.MethodGet, "В разрешённых методах нет GET")
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
				"sensor_id": 1,
				"sensor_serial_number": "1234567890",
				"timestamp": "2024-04-08T11:24:29.747Z",
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
				<SensorId>1</SensorId>
				<SensorSerialNumber>1234567890</SensorSerialNumber>
				<Timestamp>2024-04-08T11:24:29.747Z</Timestamp>
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
				"sensor_id": -1,
				"sensor_serial_number": "1234567890",
				"timestamp": "2024-04-08T11:24:29.747Z",
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
				allowed := strings.Split(w.Header().Get("Allow"), ",")
				assert.Contains(t, allowed, http.MethodOptions, "В разрешённых методах нет OPTIONS")
				assert.Contains(t, allowed, http.MethodPost, "В разрешённых методах нет POST")
			})
		}
	})
}
