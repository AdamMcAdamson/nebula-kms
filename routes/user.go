package routes

import (
	"github.com/UTDNebula/KMS/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.Engine) {

	// All routes related to users come here
	userGroup := router.Group("/user")

	// Enable CORS

	// Get User Keys
	userGroup.GET("/keys", controllers.GetUserKeys())

	// User Type
	// Privileged User Data

}
