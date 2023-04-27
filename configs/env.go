/**************************************************************************
* Environment variable configuration.
*
* This file enables developers to use a .env file to set the environment
* variables for the kms program using the godotenv package.

* Currently we are using the autoload import to automatically load the
* environment variables from the .env file.
*
* The environment variables we use are:
*  - 'MONGODB_URI' : The mongodb uri for the kms database
*  - 'Port' 	   : The port to run the server on (Default: 8080)
*
* Written by Adam Brunn (amb150230) at The University of Texas at Dallas
* for CS4485.0W1 (Nebula Platform CS Project) starting March 10, 2023.
**************************************************************************/

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
