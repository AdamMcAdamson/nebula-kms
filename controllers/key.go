package controllers

import (
	"context"
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

var keyCollection *mongo.Collection = configs.GetCollection(configs.DB, "keys")

// Create Basic Key
func CreateBasicKey() gin.HandlerFunc {
	return func(c *gin.Context) {

		var key models.Key
		var user models.User

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Pull userID from query
		userID, err := primitive.ObjectIDFromHex(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Verify userID is valid (user exists)
		err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify user has no basic keys
		var res bson.M
		err = keyCollection.FindOne(ctx, bson.D{{Key: "owner_id", Value: userID}, {Key: "key_type", Value: "Basic"}}).Decode(&res)
		if err != mongo.ErrNoDocuments && err != nil {
			// Normal Error
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		} else if err == nil {
			// User already has a basic key
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "User already has a basic key."})
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
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Update user with new key
		updateUser := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}}}, {Key: "$push", Value: bson.D{{Key: "keys", Value: key.ID}}}}

		_, err = userCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: userID}}, updateUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Return the key
		c.JSON(http.StatusCreated, responses.KeyResponse{Status: http.StatusCreated, Message: "success", Data: key})

	}
}

// Create Advanced Key
func CreateAdvancedKey() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Generate Key Name
		// name, exists := c.GetQuery("name")

		// if !exists {
		// 	rand.Seed(time.Now().Unix())
		// 	ran_str := make([]byte, 12)
		// 	for i := range ran_str {
		// 		ran_str[i] = (byte)(65 + rand.Intn(25))
		// 	}
		// 	name = "key_" + string(ran_str)
		// }

		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
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
