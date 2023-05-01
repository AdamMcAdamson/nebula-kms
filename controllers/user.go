/**************************************************************************
* User endpoint logic.
*
* These endpoints process requests related to user data within
* the developer portal of Nebula platform. Generally, these operations
* should process requests directly from the developer portal.
*
* These functions each return a gin.HandlerFunc which are called
* as descibed in routes/user.go
*
* Outside of user creation, these endpoints generally retrieve data for
* the developer portal on the user's behalf.
*
* Reponses are built using responses/user_response.go.
*
* Written by Adam Brunn (amb150230) at The University of Texas at Dallas
* for CS4485.0W1 (Nebula Platform CS Project) starting March 10, 2023.
**************************************************************************/

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

/**************************************************************************
* Get User Keys
* This returns the user's (user_id) basic and advanced keys.
**************************************************************************/
func GetUserKeys() gin.HandlerFunc {
	return func(c *gin.Context) {

		var userID primitive.ObjectID
		var userKeys map[string]interface{}

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

		// Aggregation pipeline stages
		matchOnUserID := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: userID}}}}
		lookupAdvancedKeys := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "keys"}, {Key: "localField", Value: "advanced_keys"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "advanced_keys"}}}}
		lookupBasicKey := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "keys"}, {Key: "localField", Value: "basic_key"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "basic_key"}}}}
		unwindBasicKey := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$basic_key"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}
		projectKeys := bson.D{{Key: "$project", Value: bson.D{{Key: "basic_key", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$basic_key", primitive.Null{}}}}}, {Key: "advanced_keys", Value: 1}}}}

		aggregationPipeline := bson.A{matchOnUserID, lookupAdvancedKeys, lookupBasicKey, unwindBasicKey, projectKeys}

		// Preform aggregation
		cursor, err := userCollection.Aggregate(ctx, aggregationPipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// If there is no next document, the user does not exist
		if !cursor.Next(ctx) {
			c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid user_id: User does not exist"})
			return
		}

		// Decode keys
		err = cursor.Decode(&userKeys)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Respond
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: userKeys})
	}
}

/**************************************************************************
* Get User Type
* This returns the user's (user_id) type ("Developer", "Lead", "Admin").
**************************************************************************/
func GetUserType() gin.HandlerFunc {
	return func(c *gin.Context) {

		var userID primitive.ObjectID
		var user models.User

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

		// Get user
		err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Respond
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: user.Type})
	}
}

/**************************************************************************
* Get Priviledge User Data
* This returns the priviledged data the user (user_id) has permissions
* to view concerning other users, keys, and services.
*
* Admins can see the aggregation of all services to keys to users.
* Leads can see the aggregtion of services the lead to keys to users,
* while hiding the actual keys.
**************************************************************************/
func GetPrivilegedUserData() gin.HandlerFunc {
	return func(c *gin.Context) {

		// @TODO: Extend Admin Aggregation pipeline to also grab keys for basic services

		var user models.User

		var aggregationPipeline bson.A
		var collection *mongo.Collection

		var cursor *mongo.Cursor
		var err error

		var res []map[string]interface{}
		var out interface{}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get userID from query parameter
		userID, err := primitive.ObjectIDFromHex(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Get user from database
		err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Admin Aggregation Pipeline Only
		projectServiceDetails := bson.D{{Key: "$project", Value: bson.D{{Key: "services._id", Value: "$_id"}, {Key: "services.service_name", Value: "$service_name"}, {Key: "services.service_type", Value: "$service_type"}, {Key: "services.created_at", Value: "$created_at"}, {Key: "services.updated_at", Value: "$updated_at"}, {Key: "services.source_identifiers", Value: "$source_identifiers"}}}}

		// Lead Aggregation Pipeline Only
		matchLead := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: userID}}}}
		lookupServices := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "services"}, {Key: "localField", Value: "services"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "services"}}}}
		projectServices := bson.D{{Key: "$project", Value: bson.D{{Key: "services", Value: 1}}}}
		unwindServices := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$services"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}
		// -- LookupKeys
		projectKeysLead := bson.D{{Key: "$project", Value: bson.D{{Key: "services._id", Value: 1}, {Key: "services.service_name", Value: 1}, {Key: "services.service_type", Value: 1}, {Key: "services.created_at", Value: 1}, {Key: "services.updated_at", Value: 1}, {Key: "services.source_identifiers", Value: 1}, {Key: "keys._id", Value: 1}, {Key: "keys.key", Value: "_HIDDEN_"}, {Key: "keys.key_type", Value: 1}, {Key: "keys.name", Value: 1}, {Key: "keys.owner_id", Value: 1}, {Key: "keys.service_id", Value: 1}, {Key: "keys.quota", Value: 1}, {Key: "keys.quota_type", Value: 1}, {Key: "keys.usage_remaining", Value: 1}, {Key: "keys.quota_timestamp", Value: 1}, {Key: "keys.created_at", Value: 1}, {Key: "keys.updated_at", Value: 1}, {Key: "keys.is_active", Value: 1}}}}

		// Both Lead and Admin Aggregation Pipelines
		lookupKeys := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "keys"}, {Key: "localField", Value: "services._id"}, {Key: "foreignField", Value: "service_id"}, {Key: "as", Value: "keys"}}}}
		unwindKeys := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$keys"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}
		lookupOwner := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "users"}, {Key: "localField", Value: "keys.owner_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "owner"}}}}
		projectOwner := bson.D{{Key: "$project", Value: bson.D{{Key: "services", Value: 1}, {Key: "keys", Value: 1}, {Key: "owner._id", Value: 1}, {Key: "owner.platform_user_id", Value: 1}, {Key: "owner.user_type", Value: 1}}}}
		unwindOwner := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$owner"}, {Key: "preserveNullAndEmptyArrays", Value: false}}}}
		projectOwnerIntoKey := bson.D{{Key: "$project", Value: bson.D{{Key: "services._id", Value: 1}, {Key: "services.service_name", Value: 1}, {Key: "services.service_type", Value: 1}, {Key: "services.created_at", Value: 1}, {Key: "services.updated_at", Value: 1}, {Key: "services.source_identifiers", Value: 1}, {Key: "keys._id", Value: 1}, {Key: "keys.key", Value: 1}, {Key: "keys.key_type", Value: 1}, {Key: "keys.name", Value: 1}, {Key: "keys.owner_id", Value: 1}, {Key: "keys.service_id", Value: 1}, {Key: "keys.quota", Value: 1}, {Key: "keys.quota_type", Value: 1}, {Key: "keys.usage_remaining", Value: 1}, {Key: "keys.quota_timestamp", Value: 1}, {Key: "keys.created_at", Value: 1}, {Key: "keys.updated_at", Value: 1}, {Key: "keys.is_active", Value: 1}, {Key: "keys.owner", Value: "$owner"}}}}
		groupKeys := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$services._id"}, {Key: "services", Value: bson.D{{Key: "$first", Value: "$services"}}}, {Key: "keys", Value: bson.D{{Key: "$push", Value: "$keys"}}}}}}
		projectKeysIntoService := bson.D{{Key: "$project", Value: bson.D{{Key: "services._id", Value: 1}, {Key: "services.service_name", Value: 1}, {Key: "services.service_type", Value: 1}, {Key: "services.created_at", Value: 1}, {Key: "services.updated_at", Value: 1}, {Key: "services.source_identifiers", Value: 1}, {Key: "services.keys", Value: "$keys"}}}}
		groupServices := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: primitive.Null{}}, {Key: "services", Value: bson.D{{Key: "$push", Value: "$services"}}}}}}

		// The Difference between these two aggregation pipelines is that:
		// The Admin aggregation:
		// - Is performed on the services collection
		// - Displays all services -> keys -> users
		// - Includes keys
		pipelineAdminAggregation := bson.A{projectServiceDetails, lookupKeys, unwindKeys, lookupOwner, projectOwner, unwindOwner, projectOwnerIntoKey, groupKeys, projectKeysIntoService, groupServices}

		// The Lead aggregation:
		// - Is performed on the user collection
		// - Displays led services -> keys -> users
		// - Hides keys
		pipelineLeadAggregation := bson.A{matchLead, lookupServices, projectServices, unwindServices, lookupKeys, projectKeysLead, unwindKeys, lookupOwner, projectOwner, unwindOwner, projectOwnerIntoKey, groupKeys, projectKeysIntoService, groupServices}

		// Determine course of action by checking user type
		if user.Type == "Admin" {
			aggregationPipeline = pipelineAdminAggregation
			collection = serviceCollection
		} else if user.Type == "Lead" {
			aggregationPipeline = pipelineLeadAggregation
			collection = userCollection
			if len(user.Services) == 0 { // Short circuit
				// There is no need to perform aggregation as it will return no data
				c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: nil})
				return
			}
		} else {
			c.JSON(http.StatusConflict, responses.UserResponse{Status: http.StatusConflict, Message: "error", Data: "Invalid user_id: User is not a Lead or Admin"})
			return
		}

		// Perform Aggregation
		cursor, err = collection.Aggregate(ctx, aggregationPipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}
		err = cursor.All(ctx, &res)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		if len(res) == 0 {
			// @INFO: Assumes at least one service exists
			// Constitutes an error due to handling short circuit and user type
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: "Error: No data returned from aggregation"})
			return
		} else {
			// Grab aggregation result
			out = res[0]["services"]
		}

		// Respond
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: out})
	}
}

/**************************************************************************
* Create User
* This creates a user in the kms/developer portal backend given their
* platform_user_id. This should be grabbed from Nebula SSO.
*
* Generally this should be called by the developer portal backend
* once the user wishes to access it's features.
**************************************************************************/
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
					if we.Code == 11000 { // Not unique platform_user_id
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
