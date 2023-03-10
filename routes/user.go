package routes

import (
	"github.com/gin-gonic/gin"
	// controllers
)

func UserRoute(router *gin.Engine) {

	// All routes related to users come here
	userGroup := router.Group("/user")

	// Enable CORS

	// User Keys
	// User Type
	// Privileged User Data

}
