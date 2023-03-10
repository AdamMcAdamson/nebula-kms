package controllers

import (
	"net/http"

	"github.com/UTDNebula/KMS/responses"

	"github.com/gin-gonic/gin"
)

func GetUserKeys() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.JSON(http.StatusNotImplemented, responses.UserKeysResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}
