package controllers

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/UTDNebula/kms/configs"
	"github.com/UTDNebula/kms/models"
	"github.com/UTDNebula/kms/responses"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
)

// Create Basic Key
func CreateBasicKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		/*
			POST /Create-Basic-Key
			FROM DeveloperPortalBackend
			{
				UserID
			}
			Return:
			{
				Key_Mongo_OID,
				Key,
				Timed_Quota,
				Usage_Remaining,
				Key_Created,
				Last_Modified
			}
		*/

		var key models.Key
		var user models.User

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID, err := primitive.ObjectIDFromHex(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// verify userID is valid
		err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// @TODO: verify user has no basic keys (aggregation pipeline)

		name, exists := c.GetQuery("name")

		if !exists {
			rand.Seed(time.Now().Unix())
			ran_str := make([]byte, 12)
			for i := range ran_str {
				ran_str[i] = (byte)(65 + rand.Intn(25))
			}
			name = "key_" + string(ran_str)
		}

		key.ID = primitive.NewObjectID()
		key.Type = "Basic"
		key.Name = name
		key.Quota = 200 // @TODO: Centralize values
		key.QuotaType = "Daily"
		key.UsageRemaining = key.Quota
		key.CreatedAt = time.Now()
		key.QuotaTimestamp = key.CreatedAt
		key.UpdatedAt = key.CreatedAt
		key.IsActive = true
		key.Key = configs.GenerateKey()

		// @TODO: Insert into DB

		// @TODO: Return the key

		c.JSON(http.StatusNotImplemented, responses.KeyResponse{Status: http.StatusNotImplemented, Message: "Not Implemented", Data: nil})
	}
}

// Create Advanced Key
func CreateAdvancedKey() gin.HandlerFunc {
	return func(c *gin.Context) {
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
