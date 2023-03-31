package main

import (
	"github.com/UTDNebula/kms/configs"
	"github.com/UTDNebula/kms/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	configs.InitConfig()

	// Configure Gin Router
	router := gin.Default()

	// Connect Routes
	routes.AllowedRoute(router)
	routes.KeyRoute(router)
	routes.UserRoute(router)

	// @INFO: Do not uncomment
	// routes.ServiceRoute(router)

	// Retrieve the port string to serve traffic on
	portString := configs.GetPortString()

	// Serve Traffic
	router.Run(portString)

}
