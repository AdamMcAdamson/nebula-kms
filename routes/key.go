package routes

import (
	"github.com/UTDNebula/KMS/controllers"

	"github.com/gin-gonic/gin"
)

func KeyRoute(router *gin.Engine) {

	// All routes related to keys come here
	keyGroup := router.Group("/key")

	// Enable CORS

	// Create Basic Key
	keyGroup.POST("/create-basic-key", controllers.CreateBasicKey())

	// Create Advanced Key
	// Delete Key
	// Disable Key
	// Enable Key
	// Rename Key
	// Regenerate Key
	// Set Quota for a Key

}
