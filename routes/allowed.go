package routes

import (
	"github.com/gin-gonic/gin"
	// controllers
)

func AllowedRoute(router *gin.Engine) {

	// All routes related to users come here
	allowedGroup := router.Group("/allowed")

	// Enable CORS

	// Allowed

}
