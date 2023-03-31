package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/UTDNebula/kms/models"
	"github.com/UTDNebula/kms/responses"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slices"

	"github.com/gin-gonic/gin"
)

func Allowed() gin.HandlerFunc {
	return func(c *gin.Context) {

		var key models.Key
		var service models.Service

		authKey := c.GetHeader("Authorization")
		sourceIdentifier := c.GetHeader("SourceIdentifier")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// @TODO: Handle Basic Keys (must address comments in key_model.go)

		// @TODO: Change into aggregation pipeline instead of two finds
		// Find Key
		err := keyCollection.FindOne(ctx, bson.D{{Key: "key", Value: authKey}}).Decode(&key)
		if err != nil {
			// Invalid Key
			if err == mongo.ErrNoDocuments {
				// @TODO: Make sure we want to respond like this. This means that the key is not valid, just not for this service.
				c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "Invalid Authorization Key", Data: nil, IsAllowed: false})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.AllowedResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error(), IsAllowed: false})
			return
		}

		// Find Service Key is for
		err = serviceCollection.FindOne(ctx, bson.D{{Key: "_id", Value: key.ServiceID}}).Decode(&service)
		if err != nil {
			// Unknown Service. Should never occur
			c.JSON(http.StatusInternalServerError, responses.AllowedResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error(), IsAllowed: false})
			return
		}

		// Verify valid sourceIdentifier
		if !slices.Contains(service.SourceIdentifiers, sourceIdentifier) {
			// @TODO: Make sure we want to respond like this. This means that the key is valid, just not for this service.
			c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "Invalid source-identifier", IsAllowed: false})
			return
		}

		// Authorization Granted
		c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "success", IsAllowed: true})
	}
}
