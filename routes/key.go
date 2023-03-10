package routes

import (
	"github.com/gin-gonic/gin"
	// controllers
)

func KeyRoute(router *gin.Engine) {

	// All routes related to keys come here
	keyGroup := router.Group("/key")

	// Enable CORS

	// Create Basic Key
	// Create Advanced Key
	// Delete Key
	// Disable Key
	// Enable Key
	// Rename Key
	// Regenerate Key
	// Set Quota for a Key

}
