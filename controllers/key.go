package controllers

import (
	"net/http"

	"github.com/UTDNebula/KMS/responses"

	"github.com/gin-gonic/gin"
)

// Create Basic Key
func CreateBasicKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.CreateBasicKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Create Advanced Key
func CreateAdvancedKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.CreateAdvancedKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Delete Key
func DeleteKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.DeleteKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Disable Key
func DisableKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.DisableKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Enable Key
func EnableKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.EnableKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Rename Key
func RegenerateKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.RegenerateKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Regenerate Key
func RenameKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.RenameKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Set Quota for a Key
func SetKeyQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.SetKeyQuotaResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Restore Quota for a Key
func RestoreKeyQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.SetKeyQuotaResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Change Key Holder
func ChangeKeyHolder() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.SetKeyQuotaResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Change Key Service
func ChangeKeyService() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.SetKeyQuotaResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}
