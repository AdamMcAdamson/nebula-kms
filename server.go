package main

import (
	"github.com/UTDNebula/kms/configs"
	"github.com/UTDNebula/kms/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	// Configure Gin Router
	router := gin.Default()

	// Connect Routes
	routes.AllowedRoute(router)
	routes.KeyRoute(router)
	routes.UserRoute(router)

	// Retrieve the port string to serve traffic on
	portString := configs.GetPortString()

	// Serve Traffic
	router.Run(portString)

}
