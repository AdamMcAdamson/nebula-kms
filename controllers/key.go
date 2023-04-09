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
		// @TODO: Remove debug printfs
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
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Invalid key_id: Key does not exist"})
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

		// Check if user is owner
		// If not, verify permissions
		// @INFO: We assume key.OwnerID is valid
		if key.OwnerID != userID {
			// Verify userID is valid (user exists and has permissions)
			userFilter = bson.M{"_id": userID}
			err = userCollection.FindOne(ctx, userFilter).Decode(&user)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
					return
				}
				c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
				return
			}

			// @TODO: Who should be able to disable a key (advanced or basic?) and when?
			// Check if user is an Admin, or a lead of the key's service
			// @INFO: Assumes key.ServiceID is valid
			if user.Type != "Admin" && (user.Type != "Lead" || !slices.Contains(user.Services, key.ServiceID)) {
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "The given user does not have the authority to disable this key"})
				return
			}
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

// Enable Key
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
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Invalid key_id: Key does not exist"})
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

		// Check if user is owner
		// If not, verify permissions
		// @INFO: We assume key.OwnerID is valid
		if key.OwnerID != userID {
			// Verify userID is valid (user exists and has permissions)
			userFilter = bson.M{"_id": userID}
			err = userCollection.FindOne(ctx, userFilter).Decode(&user)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Invalid recipient_user_id: User does not exist"})
					return
				}
				c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
				return
			}

			// @TODO: Who should be able to enable a key (advanced or basic?) and when?
			// Check if user is an Admin, or a lead of the key's service
			// @INFO: Assumes key.ServiceID is valid
			if user.Type != "Admin" && (user.Type != "Lead" || !slices.Contains(user.Services, key.ServiceID)) {
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "The given user does not have the authority to enable this key"})
				return
			}
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

// Rename Key
func RegenerateKey() gin.HandlerFunc {
	// PATCH /Regenerate-Key
	// FROM DeveloperPortalBackend
	// {
	//     UserID,
	//     Key_Mongo_OID,
	//     Last_Modified
	// }
	// Return:
	// {
	//     Key,
	//     Last_Modified
	// }, {409}// old Last_Modified

	return func(c *gin.Context) {
		// @Optimize: Refactor to try update in aggregation pipeline ASAP
		// and investigate reason on unsuccessful update for error reporting

		var userID primitive.ObjectID

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

		// Get Key
		err = keyCollection.FindOne(ctx, keyFilter).Decode(&key)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Invalid key_id: Key does not exist"})
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
			c.JSON(http.StatusConflict, responses.KeyResponse{Status: http.StatusConflict, Message: "error", Data: "The given user does not have the authority to regenerate this key"})
			return
		}

		// Regenerate key
		key.Key = configs.GenerateKey()
		key.UpdatedAt = time.Now().UTC()

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: key.UpdatedAt}, {Key: "key", Value: key.Key}}}}
		_, err = keyCollection.UpdateOne(ctx, keyFilter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		res := struct {
			Key       string `json:"key" bson:"key"`
			UpdatedAt string `json:"updated_at" bson:"updated_at"`
		}{
			Key:       key.Key,
			UpdatedAt: key.UpdatedAt.Format(configs.DateLayout),
		}

		// Respond with formated key.UpdatedAt time
		c.JSON(http.StatusOK, responses.KeyResponse{Status: http.StatusOK, Message: "success", Data: res})
	}
}

// Regenerate Key
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
				c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "Invalid key_id: Key does not exist"})
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

		// @TODO: Who should be able to rename a key?
		// Check if user is the owner
		// @INFO: We assume key.OwnerID is valid
		if key.OwnerID != userID {
			c.JSON(http.StatusBadRequest, responses.KeyResponse{Status: http.StatusBadRequest, Message: "error", Data: "The given user does not have the authority to rename this key"})
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
