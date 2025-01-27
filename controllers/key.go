/**************************************************************************
* Key endpoint logic.
*
* These endpoints process requests related to managing keys within
* the developer portal of Nebula platform. Generally, these operations
* should process requests directly from the developer portal.
*
* These functions each return a gin.HandlerFunc which are called
* as descibed in routes/key.go
*
* Requests which alter the state of a key generally require
* an accurate 'updated_at' timestamp, which can be acquired from a
* request to GetUserKeys in controllers/user.go.
*
* Reponses are built using responses/key_response.go.
*
* Written by Adam Brunn (amb150230) at The University of Texas at Dallas
* for CS4485.0W1 (Nebula Platform CS Project) starting March 10, 2023.
**************************************************************************/

package controllers

import (
	"context"
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

/**************************************************************************
* Create Basic Key
* This creates a basic key for the given user (user_id)
* provided they do not already have one.
**************************************************************************/
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
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid user_id"})
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
		key.QuotaNumDays = 1
		key.UsageRemaining = key.Quota
		key.CreatedAt = time.Now().UTC()
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
		updateUser := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now().UTC()}, {Key: "basic_key", Value: key.ID}}}}
		_, err = userCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: userID}}, updateUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Return the key
		c.JSON(http.StatusCreated, responses.KeyResponse{Status: http.StatusCreated, Message: "success", Data: key})

	}
}

/**************************************************************************
* Create Advanced Key
* This enables Leads and Admins (creator_user_id) to create advanced keys
* for users (recipient_user_id), provided they have the proper permissions.
*
* Admins can create advanced keys for any service (service_id).
* Leads can only create advanced keys for services they are leads for.
**************************************************************************/
func CreateAdvancedKey() gin.HandlerFunc {
	return func(c *gin.Context) {

		var key models.Key
		var creatorUser models.User
		var recipientUser models.User
		var service models.Service

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get creatorUserID
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

		// Get recipientUserID
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

		// Get serviceID
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
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid creator_user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify serviceID is valid (service exists)
		err = serviceCollection.FindOne(ctx, bson.M{"_id": serviceID}).Decode(&service)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid service_id: Service does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify creatorUser is an Admin, or a lead of the given service
		if creatorUser.Type != "Admin" && (creatorUser.Type != "Lead" || !slices.Contains(creatorUser.Services, serviceID)) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given creator_user does not have the authority to create keys for the given service"})
			return
		}

		// Verify recipientUserID is valid (user exists)
		err = userCollection.FindOne(ctx, bson.M{"_id": recipientUserID}).Decode(&recipientUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
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
		key.QuotaNumDays = 1
		key.UsageRemaining = key.Quota
		key.CreatedAt = time.Now().UTC()
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
		updateRecipientUser := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now().UTC()}}}, {Key: "$push", Value: bson.D{{Key: "advanced_keys", Value: key.ID}}}}
		_, err = userCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: recipientUserID}}, updateRecipientUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Hide the actual key and return the remaining relevant data
		key.Key = "_HIDDEN_"

		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}
		c.JSON(http.StatusCreated, responses.KeyResponse{Status: http.StatusCreated, Message: "success", Data: key})
	}
}

/**************************************************************************
* Delete Key
* This enables key owners, Leads and Admins (user_id) to delete
* advanced keys.
*
* Key owners can delete their own advanced keys.
* Admins can delete any advanced key.
* Leads can only delete advanced keys for services they are leads for.
**************************************************************************/
func DeleteKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// @Optimize: Refactor to try update in aggregation pipeline ASAP
		// and investigate reason on unsuccessful update for error reporting

		var userID primitive.ObjectID
		var userFilter bson.M
		var user models.User

		var keyID primitive.ObjectID
		var keyFilter bson.M
		var key models.Key

		var updatedAt time.Time

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get userID
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

		// Get updatedAt
		updatedAtQuery, exists := c.GetQuery("updated_at")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'updated_at' field"})
			return
		}
		updatedAt, err = time.Parse(configs.DateLayout, updatedAtQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get keyID
		keyIDQuery, exists := c.GetQuery("key_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_id' field"})
			return
		}
		keyID, err = primitive.ObjectIDFromHex(keyIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Verify keyID is valid (key exists)
		keyFilter = bson.M{"_id": keyID}

		err = keyCollection.FindOne(ctx, keyFilter).Decode(&key)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid key_id: Key does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Can only delete advanced keys
		if key.Type != "Advanced" {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Invalid key_id: Can only delete advanced keys"})
			return
		}

		// Verify matching updated_At
		if !key.UpdatedAt.Equal(updatedAt) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Out of date request: Key has been updated"})
			return
		}

		// Check if user is owner
		// If not, verify permissions
		// @INFO: We assume key.OwnerID is valid
		if key.OwnerID != userID {
			// Verify userID is valid (user exists and has permissions)
			userFilter = bson.M{"_id": userID}
			err = userCollection.FindOne(ctx, userFilter).Decode(&user)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
					return
				}
				c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
				return
			}

			// Check if user is an Admin, or a lead of the key's service
			// @INFO: Assumes key.ServiceID is valid
			if user.Type != "Admin" && (user.Type != "Lead" || !slices.Contains(user.Services, key.ServiceID)) {
				c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given user does not have the authority to delete this key"})
				return
			}
		}

		// Delete Key
		_, err = keyCollection.DeleteOne(ctx, keyFilter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Response
		c.JSON(http.StatusOK, responses.KeyResponse{Status: http.StatusOK, Message: "success", Data: nil})
	}
}

/**************************************************************************
* Disable Key
* This enables Leads and Admins (user_id) to disable keys.
*
* Admins can disable any key.
* Leads can only disable advanced keys for services they are leads for.
*
* In general, this is used when a key is compromised.
**************************************************************************/
func DisableKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// @Optimize: Refactor to try update in aggregation pipeline ASAP
		// and investigate reason on unsuccessful update for error reporting

		var userID primitive.ObjectID
		var userFilter bson.M
		var user models.User

		var keyID primitive.ObjectID
		var keyFilter bson.M
		var key models.Key

		var updatedAt time.Time

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get userID
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

		// Get updatedAt
		updatedAtQuery, exists := c.GetQuery("updated_at")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'updated_at' field"})
			return
		}
		updatedAt, err = time.Parse(configs.DateLayout, updatedAtQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get keyID
		keyIDQuery, exists := c.GetQuery("key_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_id' field"})
			return
		}
		keyID, err = primitive.ObjectIDFromHex(keyIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Verify keyID is valid (key exists)
		keyFilter = bson.M{"_id": keyID}
		err = keyCollection.FindOne(ctx, keyFilter).Decode(&key)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid key_id: Key does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify key is enabled
		if !key.IsActive {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Key is already disabled"})
			return
		}

		// Verify matching updated_At
		if !key.UpdatedAt.Equal(updatedAt) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Out of date request: Key has been updated"})
			return
		}

		// Verify userID is valid (user exists and has permissions)
		userFilter = bson.M{"_id": userID}
		err = userCollection.FindOne(ctx, userFilter).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Check if user is an Admin, or a lead of the key's service
		// @INFO: Assumes key.ServiceID is valid
		if user.Type != "Admin" && (key.Type != "Advanced" || user.Type != "Lead" || !slices.Contains(user.Services, key.ServiceID)) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given user does not have the authority to disable this key"})
			return
		}

		// Disable key
		key.IsActive = false
		key.UpdatedAt = time.Now().UTC()

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: key.UpdatedAt}, {Key: "is_active", Value: key.IsActive}}}}
		_, err = keyCollection.UpdateOne(ctx, keyFilter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Respond with formated key.UpdatedAt time
		c.JSON(http.StatusOK, responses.KeyResponse{Status: http.StatusOK, Message: "success", Data: key.UpdatedAt.Format(configs.DateLayout)})
	}
}

/**************************************************************************
* Enable Key
* This enables Leads and Admins (user_id) to (re)enable keys.
*
* Admins can enable any key.
* Leads can only enable advanced keys for services they are leads for.
*
* Key owners who lack advanced permissions and are looking to renenable
* their keys need to make requests to the endpoint for RegenerateKey().
**************************************************************************/
func EnableKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// @Optimize: Refactor to try update in aggregation pipeline ASAP
		// and investigate reason on unsuccessful update for error reporting

		var userID primitive.ObjectID
		var userFilter bson.M
		var user models.User

		var keyID primitive.ObjectID
		var keyFilter bson.M
		var key models.Key

		var updatedAt time.Time

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get userID
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

		// Get updatedAt
		updatedAtQuery, exists := c.GetQuery("updated_at")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'updated_at' field"})
			return
		}
		updatedAt, err = time.Parse(configs.DateLayout, updatedAtQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get keyID
		keyIDQuery, exists := c.GetQuery("key_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_id' field"})
			return
		}
		keyID, err = primitive.ObjectIDFromHex(keyIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Verify keyID is valid (key exists)
		keyFilter = bson.M{"_id": keyID}

		err = keyCollection.FindOne(ctx, keyFilter).Decode(&key)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid key_id: Key does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify key is disabled
		if key.IsActive {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Key is already enabled"})
			return
		}

		// Verify matching updated_At
		if !key.UpdatedAt.Equal(updatedAt) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Out of date request: Key has been updated"})
			return
		}

		// Verify userID is valid (user exists and has permissions)
		userFilter = bson.M{"_id": userID}
		err = userCollection.FindOne(ctx, userFilter).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Check if user is an Admin, or a lead of the key's service
		// @INFO: Assumes key.ServiceID is valid
		if user.Type != "Admin" && (key.Type != "Advanced" || user.Type != "Lead" || !slices.Contains(user.Services, key.ServiceID)) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given user does not have the authority to enable this key"})
			return
		}

		// Enable key
		key.IsActive = true
		key.UpdatedAt = time.Now().UTC()

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: key.UpdatedAt}, {Key: "is_active", Value: key.IsActive}}}}
		_, err = keyCollection.UpdateOne(ctx, keyFilter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Respond with formated key.UpdatedAt time
		c.JSON(http.StatusOK, responses.KeyResponse{Status: http.StatusOK, Message: "success", Data: key.UpdatedAt.Format(configs.DateLayout)})
	}
}

/**************************************************************************
* Regenerate Key
* This enables key owners, Leads and Admins (user_id) to regenerate keys.
* This also enables the key, should it had been disabled.
*
* Key owners
* Admins can regenerate any key.
* Leads can only regenerate advanced keys for services they are leads for.
**************************************************************************/
func RegenerateKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// @Optimize: Refactor to try update in aggregation pipeline ASAP
		// and investigate reason on unsuccessful update for error reporting

		var userID primitive.ObjectID
		var userFilter bson.M
		var user models.User

		var keyID primitive.ObjectID
		var keyFilter bson.M
		var key models.Key

		var updatedAt time.Time

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get userID
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

		// Get updatedAt
		updatedAtQuery, exists := c.GetQuery("updated_at")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'updated_at' field"})
			return
		}
		updatedAt, err = time.Parse(configs.DateLayout, updatedAtQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get keyID
		keyIDQuery, exists := c.GetQuery("key_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_id' field"})
			return
		}
		keyID, err = primitive.ObjectIDFromHex(keyIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get Key
		keyFilter = bson.M{"_id": keyID}

		err = keyCollection.FindOne(ctx, keyFilter).Decode(&key)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// Key does not exist
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid key_id: Key does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify matching updated_At
		if !key.UpdatedAt.Equal(updatedAt) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Out of date request: Key has been updated"})
			return
		}

		// Check if user is owner
		// @INFO: We assume key.OwnerID is valid
		if key.OwnerID != userID {
			// Verify userID is valid (user exists and has permissions)
			userFilter = bson.M{"_id": userID}
			err = userCollection.FindOne(ctx, userFilter).Decode(&user)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
					return
				}
				c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
				return
			}

			// Check if user is an Admin, or a lead of the key's service
			// @INFO: Assumes key.ServiceID is valid
			if user.Type != "Admin" && (key.Type != "Advanced" || user.Type != "Lead" || !slices.Contains(user.Services, key.ServiceID)) {
				c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given user does not have the authority to enable this key"})
				return
			}
		}

		// Regenerate key
		key.Key = configs.GenerateKey()
		key.UpdatedAt = time.Now().UTC()
		key.IsActive = true // Enable keys on regeneration

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: key.UpdatedAt}, {Key: "key", Value: key.Key}, {Key: "is_active", Value: key.IsActive}}}}
		_, err = keyCollection.UpdateOne(ctx, keyFilter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// @TODO: Refactor to key_response type
		res := struct {
			Key       string `json:"key" bson:"key"`
			UpdatedAt string `json:"updated_at" bson:"updated_at"`
			IsActive  bool   `json:"is_active" bsom:"is_active"`
		}{
			Key:       "_HIDDEN_",
			UpdatedAt: key.UpdatedAt.Format(configs.DateLayout),
			IsActive:  key.IsActive,
		}

		// Respond with formated key.UpdatedAt time
		c.JSON(http.StatusOK, responses.KeyResponse{Status: http.StatusOK, Message: "success", Data: res})
	}
}

/**************************************************************************
* Rename Key
* This enables key owners (user_id) to rename keys.
*
* Note: Leads and Admins can set key names when creating them if they wish.
**************************************************************************/
func RenameKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// @Optimize: Refactor to try update in aggregation pipeline ASAP
		// and investigate reason on unsuccessful update for error reporting

		var userID primitive.ObjectID
		// var userFilter bson.M
		// var user models.User

		var keyID primitive.ObjectID
		var keyFilter bson.M
		var key models.Key

		var keyName string

		var updatedAt time.Time

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get userID
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

		// Get updatedAt
		updatedAtQuery, exists := c.GetQuery("updated_at")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'updated_at' field"})
			return
		}
		updatedAt, err = time.Parse(configs.DateLayout, updatedAtQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get keyID
		keyIDQuery, exists := c.GetQuery("key_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_id' field"})
			return
		}
		keyID, err = primitive.ObjectIDFromHex(keyIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Verify keyID is valid (key exists)
		keyFilter = bson.M{"_id": keyID}

		err = keyCollection.FindOne(ctx, keyFilter).Decode(&key)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid key_id: Key does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Get keyName
		keyName, exists = c.GetQuery("key_name")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_name' field"})
			return
		}

		// Verify matching updated_At
		if !key.UpdatedAt.Equal(updatedAt) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Out of date request: Key has been updated"})
			return
		}

		// Check if user is the owner
		// @INFO: We assume key.OwnerID is valid
		if key.OwnerID != userID {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given user does not have the authority to rename this key"})
			return
		}

		// Rename key
		key.Name = keyName
		key.UpdatedAt = time.Now().UTC()

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: key.UpdatedAt}, {Key: "name", Value: key.Name}}}}
		_, err = keyCollection.UpdateOne(ctx, keyFilter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Respond with formated key.UpdatedAt time
		c.JSON(http.StatusOK, responses.KeyResponse{Status: http.StatusOK, Message: "success", Data: key.UpdatedAt.Format(configs.DateLayout)})
	}
}

/**************************************************************************
* Set Key Quota
* This enables Leads and Admins (user_id) to set key quotas.
*
* Admins can set key quotas for any key.
* Leads can only set key quotas of advanced keys
* for services they are leads for.
**************************************************************************/
func SetKeyQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		// @Optimize: Refactor to try update in aggregation pipeline ASAP
		// and investigate reason on unsuccessful update for error reporting

		var userID primitive.ObjectID
		var userFilter bson.M
		var user models.User

		var keyID primitive.ObjectID
		var keyFilter bson.M
		var key models.Key

		var quota int

		var updatedAt time.Time

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get userID
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

		// Get updatedAt
		updatedAtQuery, exists := c.GetQuery("updated_at")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'updated_at' field"})
			return
		}
		updatedAt, err = time.Parse(configs.DateLayout, updatedAtQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get keyID
		keyIDQuery, exists := c.GetQuery("key_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_id' field"})
			return
		}
		keyID, err = primitive.ObjectIDFromHex(keyIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get quota
		quotaStr, exists := c.GetQuery("quota")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_id' field"})
			return
		}
		quotaI64, err := strconv.ParseInt(quotaStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}
		quota = (int)(quotaI64)

		// Verify keyID is valid (key exists)
		keyFilter = bson.M{"_id": keyID}

		err = keyCollection.FindOne(ctx, keyFilter).Decode(&key)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid key_id: Key does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Get quotaNumDays (optional)
		quotaNumDaysStr, exists := c.GetQuery("quota_num_days")
		if exists {
			quotaNumDaysI64, err := strconv.ParseInt(quotaNumDaysStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
				return
			}
			// @INFO: Will only actually update if all checks are met
			key.QuotaNumDays = (int)(quotaNumDaysI64)
		}

		// Verify matching updated_At
		if !key.UpdatedAt.Equal(updatedAt) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Out of date request: Key has been updated"})
			return
		}

		// Verify userID is valid (user exists and has permissions)
		userFilter = bson.M{"_id": userID}
		err = userCollection.FindOne(ctx, userFilter).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Check if user is an Admin, or a lead of the key's service
		// @INFO: Assumes key.ServiceID is valid
		if user.Type != "Admin" && (key.Type != "Advanced" || user.Type != "Lead" || !slices.Contains(user.Services, key.ServiceID)) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given user does not have the authority to set the quota for this key"})
			return
		}

		now := time.Now().UTC()

		// Set quota
		key.Quota = quota
		key.UsageRemaining = key.Quota
		key.UpdatedAt = now
		key.QuotaTimestamp = time.Date(now.Year(), now.Month(), now.Day()+key.QuotaNumDays, 0, 0, 0, 0, time.UTC)

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: key.UpdatedAt}, {Key: "quota", Value: key.Quota}, {Key: "quota_num_days", Value: key.QuotaNumDays}, {Key: "quota_timestamp", Value: key.QuotaTimestamp}, {Key: "usage_remaining", Value: key.UsageRemaining}}}}
		_, err = keyCollection.UpdateOne(ctx, keyFilter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// @TODO: Refactor to key_response type
		res := struct {
			Quota          int    `json:"quota" bson:"quota"`
			QuotaNumDays   int    `json:"quota_num_days" bson:"quota_num_days"`
			UsageRemaining int    `json:"usage_remaining" bson:"usage_remaining"`
			QuotaTimestamp string `json:"quota_timestamp"`
			UpdatedAt      string `json:"updated_at" bson:"updated_at"`
		}{
			Quota:          key.Quota,
			QuotaNumDays:   key.QuotaNumDays,
			UsageRemaining: key.UsageRemaining,
			QuotaTimestamp: key.QuotaTimestamp.Format(configs.DateLayout),
			UpdatedAt:      key.UpdatedAt.Format(configs.DateLayout),
		}

		// Respond
		c.JSON(http.StatusOK, responses.KeyResponse{Status: http.StatusOK, Message: "success", Data: res})
	}
}

/**************************************************************************
* Restore Key Quota
* This enables Leads and Admins (user_id) to restore key quotas.
*
* Admins can restore key quotas for any key.
* Leads can only restore key quotas of advanced keys
* for services they are leads for.
*
* This does not alter the quotaTimestamp.
**************************************************************************/
func RestoreKeyQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		// @Optimize: Refactor to try update in aggregation pipeline ASAP
		// and investigate reason on unsuccessful update for error reporting

		var userID primitive.ObjectID
		var userFilter bson.M
		var user models.User

		var keyID primitive.ObjectID
		var keyFilter bson.M
		var key models.Key

		var updatedAt time.Time

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get userID
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

		// Get updatedAt
		updatedAtQuery, exists := c.GetQuery("updated_at")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'updated_at' field"})
			return
		}
		updatedAt, err = time.Parse(configs.DateLayout, updatedAtQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get keyID
		keyIDQuery, exists := c.GetQuery("key_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_id' field"})
			return
		}
		keyID, err = primitive.ObjectIDFromHex(keyIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Verify keyID is valid (key exists)
		keyFilter = bson.M{"_id": keyID}

		err = keyCollection.FindOne(ctx, keyFilter).Decode(&key)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid key_id: Key does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify matching updated_At
		if !key.UpdatedAt.Equal(updatedAt) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Out of date request: Key has been updated"})
			return
		}

		// Verify userID is valid (user exists and has permissions)
		userFilter = bson.M{"_id": userID}
		err = userCollection.FindOne(ctx, userFilter).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Check if user is an Admin, or a lead of the key's service
		// @INFO: Assumes key.ServiceID is valid
		if user.Type != "Admin" && (key.Type != "Advanced" || user.Type != "Lead" || !slices.Contains(user.Services, key.ServiceID)) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given user does not have the authority to set the quota for this key"})
			return
		}

		// Set UsageRemaining
		key.UsageRemaining = key.Quota
		key.UpdatedAt = time.Now().UTC()

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: key.UpdatedAt}, {Key: "usage_remaining", Value: key.UsageRemaining}}}}
		_, err = keyCollection.UpdateOne(ctx, keyFilter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// @TODO: Refactor to key_response type
		res := struct {
			UsageRemaining int    `json:"usage_remaining" bson:"usage_remaining"`
			UpdatedAt      string `json:"updated_at" bson:"updated_at"`
		}{
			UsageRemaining: key.UsageRemaining,
			UpdatedAt:      key.UpdatedAt.Format(configs.DateLayout),
		}

		// Respond
		c.JSON(http.StatusOK, responses.KeyResponse{Status: http.StatusOK, Message: "success", Data: res})
	}
}

/**************************************************************************
* Change Key Holder
* This enables Leads and Admins (assigner_user_id) to change
* an advanced key's owner (recipient_user_id).
*
* Admins can change the owner of any advanced key.
* Leads can only change the owner of advanced keys
* for services they are leads for.
**************************************************************************/
func ChangeKeyHolder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var key models.Key
		var assignerUser models.User
		var recipientUser models.User

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get assignerUserID
		assignerUserIDQuery, exists := c.GetQuery("assigner_user_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'assigner_user_id' field"})
			return
		}
		assignerUserID, err := primitive.ObjectIDFromHex(assignerUserIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get recipientUserID
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

		// Get KeyID
		keyIDQuery, exists := c.GetQuery("key_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_id' field"})
			return
		}
		keyID, err := primitive.ObjectIDFromHex(keyIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Verify assignerUserID is valid (user exists)
		err = userCollection.FindOne(ctx, bson.M{"_id": assignerUserID}).Decode(&assignerUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid assigner_user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify keyID is valid (key exists)
		err = keyCollection.FindOne(ctx, bson.M{"_id": keyID}).Decode(&key)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid key_id: Key does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Can only change holder of advanced keys
		if key.Type != "Advanced" {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "Invalid key_id: Can only change holder of advanced keys"})
			return
		}

		// Verify recipientUser is not the current owner
		if key.OwnerID == recipientUserID {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given recipient_user already owns this key"})
			return
		}

		// Verify assignerUser is an Admin, or a lead of the given key's service
		if assignerUser.Type != "Admin" && (assignerUser.Type != "Lead" || !slices.Contains(assignerUser.Services, key.ServiceID)) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given assigner_user does not have the authority to transfer keys for the given service"})
			return
		}

		// Verify recipientUserID is valid (user exists)
		err = userCollection.FindOne(ctx, bson.M{"_id": recipientUserID}).Decode(&recipientUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Get previous key owner
		previousKeyOwnerUserID := key.OwnerID

		// Set Key owner
		key.OwnerID = recipientUserID
		key.UpdatedAt = time.Now().UTC()

		// Update Key
		updateKey := bson.D{{Key: "$set", Value: bson.D{{Key: "owner_id", Value: key.OwnerID}, {Key: "updated_at", Value: key.UpdatedAt}}}}
		_, err = keyCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: key.ID}}, updateKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Remove key from previous owner
		updatePreviousKeyOwnerUser := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now().UTC()}}}, {Key: "$pull", Value: bson.D{{Key: "advanced_keys", Value: key.ID}}}}
		_, err = userCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: previousKeyOwnerUserID}}, updatePreviousKeyOwnerUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Update recipient user with new key
		updateRecipientUser := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now().UTC()}}}, {Key: "$push", Value: bson.D{{Key: "advanced_keys", Value: key.ID}}}}
		_, err = userCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: recipientUserID}}, updateRecipientUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// @TODO: Refactor to key_response type
		res := struct {
			KeyID     primitive.ObjectID `json:"key_id" bson:"key_id"`
			UpdatedAt string             `json:"updated_at" bson:"updated_at"`
		}{
			KeyID:     key.ID,
			UpdatedAt: key.UpdatedAt.Format(configs.DateLayout),
		}

		// Respond
		c.JSON(http.StatusOK, responses.KeyResponse{Status: http.StatusOK, Message: "success", Data: res})
	}
}

/**************************************************************************
* Change Key Service
* This enables Leads and Admins (user_id) to change a key's service.
*
* Admins can change the service of any advanced key.
* Leads can only change the service of advanced keys for services
* they are leads for, to another service they lead.
**************************************************************************/
func ChangeKeyService() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		var key models.Key
		var service models.Service

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get userID
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

		// Get keyID
		keyIDQuery, exists := c.GetQuery("key_id")
		if !exists {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Request must include the 'key_id' field"})
			return
		}
		keyID, err := primitive.ObjectIDFromHex(keyIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get serviceID
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

		// Verify userID is valid (user exists)
		err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify keyID is valid (key exists)
		err = keyCollection.FindOne(ctx, bson.M{"_id": keyID}).Decode(&key)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid key_id: Key does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify the key is an advanced key
		if key.Type != "Advanced" {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given key is a basic key. Basic keys are not for a specific service"})
			return
		}

		// Verify the given service is not the current service for the key
		if key.ServiceID == serviceID {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given key is already for the given service"})
			return
		}

		// Verify serviceID is valid (service exists)
		err = serviceCollection.FindOne(ctx, bson.M{"_id": serviceID}).Decode(&service)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid service_id: Service does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Verify user is an Admin, or a lead of the given service and the key's current service
		if user.Type != "Admin" && (user.Type != "Lead" || !(slices.Contains(user.Services, key.ServiceID) && slices.Contains(user.Services, serviceID))) {
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given user does not have the authority to transfer the given key to the given service"})
			return
		}

		// Set Key service
		key.ServiceID = serviceID
		key.UpdatedAt = time.Now().UTC()

		// Update Key
		updateKey := bson.D{{Key: "$set", Value: bson.D{{Key: "owner_id", Value: key.OwnerID}, {Key: "updated_at", Value: key.UpdatedAt}}}}
		_, err = keyCollection.UpdateOne(ctx, bson.D{{Key: "_id", Value: key.ID}}, updateKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// @TODO: Refactor to key_response type
		res := struct {
			ServiceID primitive.ObjectID `json:"serive_id" bson:"service_id"`
			UpdatedAt string             `json:"updated_at" bson:"updated_at"`
		}{
			ServiceID: key.ServiceID,
			UpdatedAt: key.UpdatedAt.Format(configs.DateLayout),
		}

		// Respond
		c.JSON(http.StatusOK, responses.KeyResponse{Status: http.StatusOK, Message: "success", Data: res})
	}
}
