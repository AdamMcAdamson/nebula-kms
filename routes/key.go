package routes

import (
	"github.com/UTDNebula/kms/controllers"

	"github.com/gin-gonic/gin"
)

func KeyRoute(router *gin.Engine) {

	// All routes related to keys come here
	keyGroup := router.Group("/key")

	// Enable CORS (?)

	// Create Basic Key
	keyGroup.POST("/create-basic-key", controllers.CreateBasicKey())

	// Create Advanced Key
	keyGroup.POST("/create-advanced-key", controllers.CreateAdvancedKey())

	// Delete Key
	keyGroup.DELETE("/delete", controllers.DeleteKey())

	// Disable Key
	keyGroup.PATCH("/disable", controllers.DisableKey())

	// Enable Key
	keyGroup.PATCH("/enable", controllers.EnableKey())

	// Rename Key
	keyGroup.PATCH("/regenerate", controllers.RegenerateKey())

	// Regenerate Key
	keyGroup.PATCH("/rename", controllers.RenameKey())

	// Set Quota for a Key
	keyGroup.PATCH("/set-quota", controllers.SetKeyQuota())

	// Restore Quota for a Key
	keyGroup.PATCH("/restore-quota", controllers.RestoreKeyQuota())

	// Change Key Holder
	keyGroup.PATCH("/change-holder", controllers.ChangeKeyHolder())

	// Change Key Service
	keyGroup.PATCH("/change-service", controllers.ChangeKeyService())
}
