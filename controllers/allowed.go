/**************************************************************************
* Allowed endpoint logic.
*
* This validates requests to access service APIs provided by Nebula Labs
* with a given authorization key and source identifier.
*
* This is achieved by finding the key document for the given 'authKey'
* from our database, then performing a series of checks to determine if
* access should be granted.
*
* The following request headers are required:
* 'Authorization' - The authorization key.
* 'Requested-service' - The source identifier of the requested service.
*
* Should the key be valid, be active, have usage remaining, and be for
* the requested service, then access should be granted.
*
* The 'IsAllowed' field of the response informs whether the request
* should be granted.
*
* NOTE: Basic keys are for any service of service type 'Basic' while
*       Advanced keys are for a specific service.
*
* Written by Adam Brunn (amb150230) at The University of Texas at Dallas
* for CS4485.0W1 (Nebula Platform CS Project) starting March 10, 2023.
**************************************************************************/

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

/**************************************************************************
* Allowed endpoint function as described above. This returns a
* gin.HandlerFunc which is called as descibed in routes/allowed.go
**************************************************************************/
func Allowed() gin.HandlerFunc {
	return func(c *gin.Context) {

		var key models.Key
		var service models.Service

		authKey := c.GetHeader("Authorization")
		sourceIdentifier := c.GetHeader("Requested-service")

		// Missing required headers
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

		// Key has no usage remaining
		if key.UsageRemaining <= 0 {
			c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "Quota reached", IsAllowed: false})
			return
		}

		// Key is not active
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

		// Update key's usage remaining
		key.UsageRemaining -= 1
		updateKey := bson.D{{Key: "$set", Value: bson.D{{Key: "usage_remaining", Value: key.UsageRemaining}, {Key: "last_used", Value: time.Now()}}}}
		_, err = keyCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: key.ID}}, updateKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.AllowedResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error(), IsAllowed: false})
			return
		}

		// Authorization Granted
		c.JSON(http.StatusOK, responses.AllowedResponse{Status: http.StatusOK, Message: "success", IsAllowed: true})
	}
}
