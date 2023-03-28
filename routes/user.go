package routes

import (
	"github.com/UTDNebula/kms/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.Engine) {

	// All routes related to users come here
	userGroup := router.Group("/user")

	// Enable CORS (?)

	// Get User Keys
	userGroup.GET("/keys", controllers.GetUserKeys())

	// User Type
	userGroup.GET("/type", controllers.GetUserType())

	// Privileged User Data
	userGroup.GET("/priviledged-data", controllers.GetPrivilegedUserData())

	// Create User
	userGroup.POST("/create", controllers.CreateUser())

}
