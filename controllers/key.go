package controllers

import (
	"net/http"

	"github.com/UTDNebula/KMS/responses"

	"github.com/gin-gonic/gin"
)

func CreateBasicKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.CreateBasicKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

func CreateAdvancedKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.CreateAdvancedKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

func DeleteKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.DeleteKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

func DisableKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.DisableKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

func EnableKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.EnableKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

func RegenerateKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.RegenerateKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

func RenameKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.RenameKeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

func SetKeyQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.SetKeyQuotaResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}
