package http

import (
	"github.com/gin-gonic/gin"
)

func setupRouter(r *gin.RouterGroup, uc UseCases) {
	setupEventsHandler(r.Group("/events"), uc)
	setupSensorsHandler(r.Group("/sensors"), uc)
	setupUsersHandler(r.Group("/users"), uc)
}
