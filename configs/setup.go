package configs

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitConfig() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func ConnectDB() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(GetEnvMongoURI()))
	if err != nil {
		log.Fatalf("Unable to create MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	fmt.Println("Connected to MongoDB")
	return client
}

// Client instance
var DB *mongo.Client = ConnectDB()

// getting database collections
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("kmsDB").Collection(collectionName)
	return collection
}

// creates a goroutine for refreshing keys' quotaTimestamp
func RefreshUsageRemainingGoroutine() {
	RefreshUsageRemainingOperation()

	// Calculate the duration until the next midnight UTC
	now := time.Now().UTC()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	duration := time.Until(nextMidnight)

	// Create a ticker with the duration until the next midnight
	ticker := time.NewTicker(duration)

	go func() {
		for {
			<-ticker.C // Wait for the ticker to fire

			// RefreshUsageRemainingOperation will be triggered every day exactly at midnight UTC
			RefreshUsageRemainingOperation()
			fmt.Println("Daily Quotas Refreshing...")

			// Reset the ticker for the next day
			now := time.Now().UTC()
			nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
			duration := time.Until(nextMidnight)
			ticker.Reset(duration)
		}
	}()
}
