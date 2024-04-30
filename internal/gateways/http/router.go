package http

import (
	"github.com/gin-gonic/gin"
)

func setupRouter(r gin.IRouter, uc UseCases) {
	if rEngine, ok := r.(*gin.Engine); ok {
		rEngine.HandleMethodNotAllowed = true // Hack (bad tests design):
		// tests require to set "/api" group up *at the *level above the router*,
		// and also require to set the 405 response up *at the router level*,
		// that is impossible as gin.RouterGroup doesn't have that flag
	}

	setupEventsHandler(r.Group("/events"), uc)
	setupSensorsHandler(r.Group("/sensors"), uc)
	setupUsersHandler(r.Group("/users"), uc)
}
