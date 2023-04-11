package controllers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/UTDNebula/kms/configs"
	"github.com/UTDNebula/kms/models"
	"github.com/UTDNebula/kms/responses"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")

func GetUserKeys() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var user models.User
		var userKeys map[string]interface{}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID, err := primitive.ObjectIDFromHex(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		pipelineMatchOnUserID := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: userID}}}}
		pipelineLookupAdvancedKeys := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "keys"}, {Key: "localField", Value: "advanced_keys"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "advanced_keys"}}}}
		pipelineLookupBasicKey := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "keys"}, {Key: "localField", Value: "basic_key"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "basic_key"}}}}
		pipelineUnwindBasicKey := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$basic_key"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}
		pipelineProjectKeyFields := bson.D{{Key: "$project", Value: bson.D{{Key: "basic_key", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$basic_key", primitive.Null{}}}}}, {Key: "advanced_keys", Value: 1}}}}

		mongoPipeline := bson.A{pipelineMatchOnUserID, pipelineLookupAdvancedKeys, pipelineLookupBasicKey, pipelineUnwindBasicKey, pipelineProjectKeyFields}

		cursor, err := userCollection.Aggregate(ctx, mongoPipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}
		if !cursor.Next(ctx) {
			c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid user_id: User does not exist"})
			return
		}
		err = cursor.Decode(&userKeys)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: userKeys})
	}
}

func GetUserType() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user models.User

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		err = userCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: user.Type})
	}
}

func GetPrivilegedUserData() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.UserResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		var newUser models.User

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Grab new user data from request body
		if err := c.BindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// @TODO: Call Platform API to verify existence of platform user
		if newUser.PlatformUserID == (primitive.ObjectID{}) {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include platform_user_id"})
			return
		}

		// Build newUser
		newUser.ID = primitive.NewObjectID()
		newUser.Type = "Developer"
		newUser.CreatedAt = time.Now()
		newUser.UpdatedAt = newUser.CreatedAt
		newUser.BasicKey = primitive.NilObjectID
		newUser.AdvancedKeys = []primitive.ObjectID{}
		newUser.Services = nil

		// Insert newUser into the database
		_, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			// Give clean error response on attempt to create duplicate profile for one platform user
			var e mongo.WriteException
			if errors.As(err, &e) {
				for _, we := range e.WriteErrors {
					if we.Code == 11000 {
						c.JSON(http.StatusConflict, responses.UserResponse{Status: http.StatusConflict, Message: "error", Data: "The given platform_user_id is already associated with a KMS user"})
						return
					}
				}
			}
			// Other errors
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Return newUser
		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: newUser})
	}
}
