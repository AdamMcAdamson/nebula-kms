package routes

import (
	"github.com/UTDNebula/KMS/controllers"

	"github.com/gin-gonic/gin"
)

func AllowedRoute(router *gin.Engine) {

	// All KMS Keys are verified through the allowed endpoint
	router.GET("/allowed", controllers.Allowed())

}
