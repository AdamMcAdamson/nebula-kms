/**************************************************************************
* Service endpoint logic.
*
* This enables the creation of services in the Nebula Labs
* kms/developer portal backend.
*
* Currently this should not be live in the kms deployment,
* and strictly serves as a tool for creating one-off services
* when running locally to ease the kms Admin experience.
*
* Once the Admin page of developer portal is fleshed-out and the front-end
* supports the creation of kms services, this can then be leveraged,
* and be a part of the live kms deployment for Nebula Labs.
*
* Reponses are built using responses/service_response.go.
*
* Written by Adam Brunn (amb150230) at The University of Texas at Dallas
* for CS4485.0W1 (Nebula Platform CS Project) starting March 10, 2023.
**************************************************************************/

package controllers

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/UTDNebula/kms/configs"
	"github.com/UTDNebula/kms/models"
	"github.com/UTDNebula/kms/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

var serviceCollection *mongo.Collection = configs.GetCollection(configs.DB, "services")

/**************************************************************************
* Create Service function as described above. This returns a
* gin.HandlerFunc which is called as descibed in routes/service.go
**************************************************************************/
func CreateService() gin.HandlerFunc {
	return func(c *gin.Context) {

		var newService models.Service

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Grab newService data from request body
		if err := c.BindJSON(&newService); err != nil {
			c.JSON(http.StatusBadRequest, responses.ServiceResponse{Status: http.StatusBadRequest, Message: "error", Data: err.Error()})
			return
		}

		// Verify valid service type
		if newService.Type != "PublicProduction" && newService.Type != "PrivateProduction" && newService.Type != "Staging" {
			c.JSON(http.StatusConflict, responses.ServiceResponse{Status: http.StatusConflict, Message: "error", Data: "Invald service_type. Must be 'PublicProduction', 'PrivateProduction', or 'Staging'"})
			return
		}

		// Generate Service Name
		if newService.Name == "" {
			rand.Seed(time.Now().Unix())
			ran_str := make([]byte, 12)
			for i := range ran_str {
				ran_str[i] = (byte)(65 + rand.Intn(25))
			}
			newService.Name = "service_" + string(ran_str)
		}

		// Ensure sourceIdenitifiers is not nil
		if newService.SourceIdentifiers == nil {
			newService.SourceIdentifiers = []string{}
		}

		// Create newService
		newService.ID = primitive.NewObjectID()
		newService.CreatedAt = time.Now()
		newService.UpdatedAt = newService.CreatedAt

		// Insert newService into the database
		_, err := serviceCollection.InsertOne(ctx, newService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.ServiceResponse{Status: http.StatusInternalServerError, Message: "error", Data: err.Error()})
			return
		}

		// Respond
		c.JSON(http.StatusCreated, responses.ServiceResponse{Status: http.StatusCreated, Message: "success", Data: newService})
	}
}
