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
		sourceIdentifier := c.GetHeader("Requested-service")

		if authKey == "" {
			c.JSON(http.StatusBadRequest, responses.AllowedResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include Authorization header", IsAllowed: false})
			return
		}

		if sourceIdentifier == "" {
			c.JSON(http.StatusBadRequest, responses.AllowedResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include Requested-service header", IsAllowed: false})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// @TODO: Change into aggregation pipeline instead of two finds and an update (keep in mind there are two potential paths basic on key.Type)

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

		if key.UsageRemaining <= 0 {
			c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "Quota reached", IsAllowed: false})
			return
		}

		if !key.IsActive {
			c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "Key is disabled", IsAllowed: false})
			return
		}

		if key.Type == "Basic" {
			// Basic Key
			// Here we check for an overlap between the basic services and the sourceIdentifier.
			// Since we already know the key is valid for all basic services,
			// we just need to check if they are requesting a valid basic service.
			err = serviceCollection.FindOne(ctx, bson.D{{Key: "service_type", Value: "Basic"}, {Key: "source_identifiers", Value: sourceIdentifier}}).Decode(&service)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "Invalid Requested-service", IsAllowed: false})
					return
				}
				c.JSON(http.StatusInternalServerError, responses.AllowedResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error(), IsAllowed: false})
				return
			}
		} else {
			// Advanced Key
			// Find Service Key is for
			err = serviceCollection.FindOne(ctx, bson.D{{Key: "_id", Value: key.ServiceID}}).Decode(&service)
			if err != nil {
				c.JSON(http.StatusInternalServerError, responses.AllowedResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error(), IsAllowed: false})
				return
			}

			// Verify valid sourceIdentifier
			if !slices.Contains(service.SourceIdentifiers, sourceIdentifier) {
				// @TODO: Make sure we want to respond like this. This means that the key is valid, just not for this service.
				c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "Invalid Requested-service", IsAllowed: false})
				return
			}
		}

		key.UsageRemaining -= 1

		// @TODO: Implement logging and push this to after we tell the gateway that the key is allowed
		// Update key's usage remaining
		// updateKey := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}, {Key: "usage_remaining", Value: key.UsageRemaining}}}}
		updateKey := bson.D{{Key: "$set", Value: bson.D{{Key: "usage_remaining", Value: key.UsageRemaining}, {Key: "last_used", Value: time.Now()}}}}
		_, err = keyCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: key.ID}}, updateKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.AllowedResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error(), IsAllowed: false})
			return
		}

		// @TODO: Implement logging and push this to before we update the key's information
		// Authorization Granted
		c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "success", IsAllowed: true})
	}
}
