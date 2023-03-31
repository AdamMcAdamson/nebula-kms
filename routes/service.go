package routes

import (
	"github.com/UTDNebula/kms/controllers"

	"github.com/gin-gonic/gin"
)

func ServiceRoute(router *gin.Engine) {

	// All routes related to services come here
	serviceGroup := router.Group("/service")

	// All KMS Keys are verified through the allowed endpoint
	serviceGroup.POST("/create", controllers.CreateService())

}
