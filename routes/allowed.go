package routes

import (
	"github.com/UTDNebula/kms/controllers"

	"github.com/gin-gonic/gin"
)

func AllowedRoute(router *gin.Engine) {

	// All KMS Keys are verified through the allowed endpoint
	router.GET("/allowed", controllers.Allowed())

}
