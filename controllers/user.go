package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/UTDNebula/kms/configs"
	"github.com/UTDNebula/kms/models"
	"github.com/UTDNebula/kms/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")

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

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var newUser models.User

		if err := c.BindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// @TODO: Call Platform API to verify existence of platform user
		if newUser.PlatformUserID == (primitive.ObjectID{}) {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include platform_user_id"})
			return
		}

		newUser.ID = primitive.NewObjectID()
		newUser.Type = "Developer"
		newUser.CreatedAt = time.Now()
		newUser.UpdatedAt = newUser.CreatedAt
		newUser.Keys = []primitive.ObjectID{}
		newUser.Services = nil

		_, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: newUser})
	}
}
