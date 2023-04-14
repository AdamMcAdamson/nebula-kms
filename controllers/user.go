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

func GetUserKeys() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var user models.User
		var userKeys map[string]interface{}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID, err := primitive.ObjectIDFromHex(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		pipelineMatchOnUserID := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: userID}}}}
		pipelineLookupAdvancedKeys := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "keys"}, {Key: "localField", Value: "advanced_keys"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "advanced_keys"}}}}
		pipelineLookupBasicKey := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "keys"}, {Key: "localField", Value: "basic_key"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "basic_key"}}}}
		pipelineUnwindBasicKey := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$basic_key"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}
		pipelineProjectKeyFields := bson.D{{Key: "$project", Value: bson.D{{Key: "basic_key", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$basic_key", primitive.Null{}}}}}, {Key: "advanced_keys", Value: 1}}}}

		mongoPipeline := bson.A{pipelineMatchOnUserID, pipelineLookupAdvancedKeys, pipelineLookupBasicKey, pipelineUnwindBasicKey, pipelineProjectKeyFields}

		cursor, err := userCollection.Aggregate(ctx, mongoPipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}
		if !cursor.Next(ctx) {
			c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid user_id: User does not exist"})
			return
		}
		err = cursor.Decode(&userKeys)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.KeyResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: userKeys})
	}
}

func GetUserType() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user models.User

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID, err := primitive.ObjectIDFromHex(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: user.Type})
	}
}

func GetPrivilegedUserData() gin.HandlerFunc {
<<<<<<< Updated upstream
=======
<<<<<<< Updated upstream
=======
>>>>>>> Stashed changes
	// GET /KMS-Privileged-Data
	// FROM DeveloperPortalBackend
	// {
	// 	UserID
	// }
	// Return:
	// {
	// 	User_Collection * Service (Filtered by User Type) * Key_Collection (Limited)
	// }
<<<<<<< Updated upstream
=======

	// @TODO: Admins can see key values, Leads cannot
>>>>>>> Stashed changes
>>>>>>> Stashed changes
	return func(c *gin.Context) {

		var user models.User
		var aggregationPipeline bson.A
		var cursor *mongo.Cursor
		var err error
		var res map[string]interface{}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID, err := primitive.ObjectIDFromHex(c.Query("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, responses.KeyResponse{Status: http.StatusNotFound, Message: "error", Data: "Invalid user_id: User does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Admin Aggregation Only
		projectServiceDetails := bson.D{{Key: "$project", Value: bson.D{{Key: "services._id", Value: "$_id"}, {Key: "services.service_name", Value: "$service_name"}, {Key: "services.service_type", Value: "$service_type"}, {Key: "services.created_at", Value: "$created_at"}, {Key: "services.updated_at", Value: "$updated_at"}, {Key: "services.source_identifiers", Value: "$source_identifiers"}}}}

		// Lead Aggregation Only
		matchLead := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: userID}}}}
		lookupServices := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "services"}, {Key: "localField", Value: "services"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "services"}}}}
		projectServices := bson.D{{Key: "$project", Value: bson.D{{Key: "services", Value: 1}}}}
		unwindServices := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$services"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

		// Both Lead and Admin Aggregation
		lookupKeys := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "keys"}, {Key: "localField", Value: "services._id"}, {Key: "foreignField", Value: "service_id"}, {Key: "as", Value: "keys"}}}}
		projectKeys := bson.D{{Key: "$project", Value: bson.D{{Key: "keys.key", Value: 0}}}}
		unwindKeys := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$keys"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}
		lookupOwner := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "users"}, {Key: "localField", Value: "keys.owner_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "owner"}}}}
		projectOwner := bson.D{{Key: "$project", Value: bson.D{{Key: "owner.basic_key", Value: 0}, {Key: "owner.created_at", Value: 0}, {Key: "owner.updated_at", Value: 0}, {Key: "owner.basic_key", Value: 0}, {Key: "owner.advanced_keys", Value: 0}}}}
		unwindOwner := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$owner"}, {Key: "preserveNullAndEmptyArrays", Value: false}}}}
		projectOwnerIntoKey := bson.D{{Key: "$project", Value: bson.D{{Key: "services._id", Value: 1}, {Key: "services.service_name", Value: 1}, {Key: "services.service_type", Value: 1}, {Key: "services.created_at", Value: 1}, {Key: "services.updated_at", Value: 1}, {Key: "services.source_identifiers", Value: 1}, {Key: "keys._id", Value: 1}, {Key: "keys.key_type", Value: 1}, {Key: "keys.name", Value: 1}, {Key: "keys.owner_id", Value: 1}, {Key: "keys.service_id", Value: 1}, {Key: "keys.quota", Value: 1}, {Key: "keys.quota_type", Value: 1}, {Key: "keys.usage_remaining", Value: 1}, {Key: "keys.quota_timestamp", Value: 1}, {Key: "keys.created_at", Value: 1}, {Key: "keys.updated_at", Value: 1}, {Key: "keys.is_active", Value: 1}, {Key: "keys.owner", Value: "$owner"}}}}
		groupKeys := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$services._id"}, {Key: "services", Value: bson.D{{Key: "$first", Value: "$services"}}}, {Key: "keys", Value: bson.D{{Key: "$push", Value: "$keys"}}}}}}
		projectKeysIntoService := bson.D{{Key: "$project", Value: bson.D{{Key: "services._id", Value: 1}, {Key: "services.service_name", Value: 1}, {Key: "services.service_type", Value: 1}, {Key: "services.created_at", Value: 1}, {Key: "services.updated_at", Value: 1}, {Key: "services.source_identifiers", Value: 1}, {Key: "services.keys", Value: "$keys"}}}}
		groupServices := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: primitive.Null{}}, {Key: "services", Value: bson.D{{Key: "$push", Value: "$services"}}}}}}

		pipelineLeadAggregation := bson.A{matchLead, lookupServices, projectServices, unwindServices, lookupKeys, projectKeys, unwindKeys, lookupOwner, projectOwner, unwindOwner, projectOwnerIntoKey, groupKeys, projectKeysIntoService, groupServices}

		pipelineAdminAggregation := bson.A{projectServiceDetails, lookupKeys, projectKeys, unwindKeys, lookupOwner, projectOwner, unwindOwner, projectOwnerIntoKey, groupKeys, projectKeysIntoService, groupServices}

		// Determine course of action by checking user type
		if user.Type == "Admin" {
			aggregationPipeline = pipelineAdminAggregation
		} else if user.Type == "Lead" {
			aggregationPipeline = pipelineLeadAggregation
			if len(user.Services) == 0 {
				c.JSON(http.StatusConflict, responses.UserResponse{Status: http.StatusConflict, Message: "error", Data: "Invalid user_id: User is not a Lead or Admin"})
				return
			}
		} else {
			c.JSON(http.StatusConflict, responses.UserResponse{Status: http.StatusConflict, Message: "error", Data: "Invalid user_id: User is not a Lead or Admin"})
			return
		}

		cursor, err = serviceCollection.Aggregate(ctx, aggregationPipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}
		cursor.All(ctx, res)
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: res["services"]})
	}
}

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
					if we.Code == 11000 {
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
