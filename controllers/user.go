package controllers

import (
	"net/http"

	"github.com/UTDNebula/kms/responses"

	"github.com/gin-gonic/gin"
)

func GetUserKeys() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.UserResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

func GetUserType() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.UserResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

func GetPrivilegedUserData() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.UserResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}
