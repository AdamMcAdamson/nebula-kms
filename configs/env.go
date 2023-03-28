package configs

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

func GetPortString() string {

	portNumber, exist := os.LookupEnv("Port")
	if !exist {
		portNumber = "8080"
	}

	portString := fmt.Sprintf(":%s", portNumber)

	return portString
}

func GetEnvMongoURI() string {

	uri, exist := os.LookupEnv("MONGODB_URI")
	if !exist {
		log.Fatalf("Error loading 'MONGODB_URI' from the .env file")
	}

	return uri
}
