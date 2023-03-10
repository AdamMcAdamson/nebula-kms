package main

import (
	"github.com/UTDNebula/KMS/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	// Establish the connection to the database
	//configs.ConnectDB()

	// Configure Gin Router
	router := gin.Default()

	// Connect Routes
	routes.AllowedRoute(router)
	routes.KeyRoute(router)
	routes.UserRoute(router)

	// Retrieve the port string to serve traffic on
	//portString := configs.GetPortString()

	// Serve Traffic
	const portString string = ":8080"
	router.Run(portString)

}
