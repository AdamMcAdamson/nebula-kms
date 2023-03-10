package controllers

import (
	"net/http"

	"github.com/UTDNebula/KMS/responses"

	"github.com/gin-gonic/gin"
)

func CreateBasicKey() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.JSON(http.StatusOK, responses.CreateBasicKeyResponse{Status: http.StatusOK, Message: "success", Data: nil})
	}
}
