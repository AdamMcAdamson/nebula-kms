package main

import (
	"github.com/gin-gonic/gin"
)

func main() {

	// Establish the connection to the database
	//configs.ConnectDB()

	// Configure Gin Router
	router := gin.Default()

	// Connect Routes
	// routes.CourseRoute(router)
	// routes.DegreeRoute(router)
	// routes.ExamRoute(router)
	// routes.SectionRoute(router)
	// routes.ProfessorRoute(router)

	// Retrieve the port string to serve traffic on
	//portString := configs.GetPortString()

	// Serve Traffic
	const portString string = ":8080"
	router.Run(portString)

}
