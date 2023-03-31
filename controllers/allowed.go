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

		// @TODO: Change into aggregation pipeline instead of two finds and an update
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

		if key.UsageRemaining == 0 {
			c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "Quota reached", IsAllowed: false})
			return
		}

		if !key.IsActive {
			c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "Key is disabled", IsAllowed: false})
			return
		}

		// Find Service Key is for
		err = serviceCollection.FindOne(ctx, bson.D{{Key: "_id", Value: key.ServiceID}}).Decode(&service)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.AllowedResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error(), IsAllowed: false})
			return
		}

		// Verify valid sourceIdentifier
		if !slices.Contains(service.SourceIdentifiers, sourceIdentifier) {
			// @TODO: Make sure we want to respond like this. This means that the key is valid, just not for this service.
			c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "Invalid source-identifier", IsAllowed: false})
			return
		}

		key.UsageRemaining -= 1

		// Update key's usage remaining
		updateKey := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}, {Key: "usage_remaining", Value: key.UsageRemaining}}}}
		_, err = keyCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: key.ID}}, updateKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.AllowedResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error(), IsAllowed: false})
			return
		}

		// @TODO: Decide if we want to move this before we update the key's usage remaining (assume it will be successful) to minimize latency
		// Authorization Granted
		c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "success", IsAllowed: true})
	}
}
