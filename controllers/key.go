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
