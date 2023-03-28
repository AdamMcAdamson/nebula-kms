package controllers

import (
	"net/http"

	"github.com/UTDNebula/kms/responses"

	"github.com/gin-gonic/gin"
)

func Allowed() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.AllowedResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}
