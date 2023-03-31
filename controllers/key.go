package controllers

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/UTDNebula/kms/configs"
	"github.com/UTDNebula/kms/models"
	"github.com/UTDNebula/kms/responses"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slices"

	"github.com/gin-gonic/gin"
)

var keyCollection *mongo.Collection = configs.GetCollection(configs.DB, "keys")

// Create Basic Key
func CreateBasicKey() gin.HandlerFunc {
	return func(c *gin.Context) {

		var key models.Key
		var user models.User

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Pull userID from query
		userIDQuery, exists := c.GetQuery("user_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'user_id' field"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(userIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Verify userID is valid (user exists)
		err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Invalid user_id"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify user has no basic key
		if user.BasicKey != primitive.NilObjectID {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "User already has a basic key"})
			return
		}

		// Build Key
		key.ID = primitive.NewObjectID()
		key.Type = "Basic"
		key.Name = "Basic_Key"
		key.OwnerID = userID
		key.Quota = 100 // @TODO: Centralize const values
		key.QuotaType = "Daily"
		key.UsageRemaining = key.Quota
		key.CreatedAt = time.Now()
		key.QuotaTimestamp = key.CreatedAt
		key.UpdatedAt = key.CreatedAt
		key.IsActive = true
		key.Key = configs.GenerateKey()

		// Create Key
		_, err = keyCollection.InsertOne(ctx, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Update user with new key
		updateUser := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}, {Key: "basic_key", Value: key.ID}}}}
		_, err = userCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: userID}}, updateUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Return the key
		c.JSON(http.StatusCreated, responses.KeyResponse{Status: http.StatusCreated, Message: "success", Data: key})

	}
}

// Create Advanced Key
func CreateAdvancedKey() gin.HandlerFunc {
	return func(c *gin.Context) {

		var key models.Key
		var creatorUser models.User
		var recipientUser models.User
		var service models.Service

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Pull objectID fields from query
		creatorUserIDQuery, exists := c.GetQuery("creator_user_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'creator_user_id' field"})
			return
		}
		creatorUserID, err := primitive.ObjectIDFromHex(creatorUserIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		recipientUserIDQuery, exists := c.GetQuery("recipient_user_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'recipient_user_id' field"})
			return
		}
		recipientUserID, err := primitive.ObjectIDFromHex(recipientUserIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		serviceIDQuery, exists := c.GetQuery("service_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'service_id' field"})
			return
		}
		serviceID, err := primitive.ObjectIDFromHex(serviceIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Verify creatorUserID is valid (user exists)
		err = userCollection.FindOne(ctx, bson.M{"_id": creatorUserID}).Decode(&creatorUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Invalid creator_user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify serviceID is valid (service exists)
		err = serviceCollection.FindOne(ctx, bson.M{"_id": serviceID}).Decode(&service)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Invalid service_id: Service does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify creatorUser is an Admin, or a lead of the given service
		if creatorUser.Type != "Admin" && (creatorUser.Type != "Lead" || !slices.Contains(creatorUser.Services, serviceID)) {
			println(creatorUser.Type + ": ")
			log.Printf("%v", creatorUser.Services)
			print(" : ")
			println(serviceID.String())
			println(slices.Contains(creatorUser.Services, serviceID))

			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "The given creator_user does not have the authority to create keys for the given service"})
			return
		}

		// Verify recipientUserID is valid (user exists)
		err = userCollection.FindOne(ctx, bson.M{"_id": recipientUserID}).Decode(&recipientUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Grab remaining query fields
		key.Name = c.Query("key_name")
		key.Quota, err = strconv.Atoi(c.Query("quota"))

		// Set quota if error or not given
		if err != nil || key.Quota == 0 {
			key.Quota = 1000 // @TODO: Centralize const values
		}

		// Generate key name if not given
		if key.Name == "" {
			ran_str := make([]byte, 12)
			for i := range ran_str {
				ran_str[i] = (byte)(65 + rand.Intn(25))
			}
			key.Name = "key_" + string(ran_str)
		}

		// Build Key
		key.ID = primitive.NewObjectID()
		key.Type = "Advanced"
		key.OwnerID = recipientUserID
		key.ServiceID = serviceID
		key.QuotaType = "Daily" // @TODO: Handle alternative quota types
		key.UsageRemaining = key.Quota
		key.CreatedAt = time.Now()
		key.QuotaTimestamp = key.CreatedAt
		key.UpdatedAt = key.CreatedAt
		key.IsActive = true
		key.Key = configs.GenerateKey()

		// Create Key
		_, err = keyCollection.InsertOne(ctx, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Update recipient user with new key
		updateRecipientUser := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}}}, {Key: "$push", Value: bson.D{{Key: "advanced_keys", Value: key.ID}}}}
		_, err = userCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: recipientUserID}}, updateRecipientUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Hide the actual key and return the remaining relevant data
		key.Key = "_HIDDEN_"
		c.JSON(http.StatusCreated, responses.KeyResponse{Status: http.StatusCreated, Message: "success", Data: key})
	}
}

// Delete Key
func DeleteKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Disable Key
func DisableKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Enable Key
func EnableKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Rename Key
func RegenerateKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Regenerate Key
func RenameKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Set Quota for a Key
func SetKeyQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Restore Quota for a Key
func RestoreKeyQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Change Key Holder
func ChangeKeyHolder() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Change Key Service
func ChangeKeyService() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}
