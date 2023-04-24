package configs

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// Layout for parsing Bson Date strings
const DateLayout = "2006-01-02T15:04:05.000-07:00"

// Key Generation
const key_length = 128
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits
)

func GenerateKey() string {
	// From https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
	sb := strings.Builder{}
	sb.Grow(key_length)
	for i, cache, remain := key_length-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func RefreshUsageRemainingOperation() {

	keyCollection := GetCollection(DB, "keys")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchQuotaTimestamps := bson.D{{Key: "$match", Value: bson.D{{Key: "quota_timestamp", Value: bson.D{{Key: "$lt", Value: time.Now()}}}}}}
	setQuotaDetails := bson.D{{Key: "$set", Value: bson.D{{Key: "usage_remaining", Value: "$quota"}, {Key: "updated_at", Value: time.Now()}, {Key: "quota_timestamp", Value: bson.D{{Key: "$dateAdd", Value: bson.D{{Key: "startDate", Value: bson.D{{Key: "$dateTrunc", Value: bson.D{{Key: "date", Value: time.Now()}, {Key: "unit", Value: "day"}}}}}, {Key: "unit", Value: "day"}, {Key: "amount", Value: "$quota_num_days"}}}}}}}}
	mergeToKeysCollection := bson.D{{Key: "$merge", Value: "keys"}}

	refreshQuotaPipeline := bson.A{matchQuotaTimestamps, setQuotaDetails, mergeToKeysCollection}

	_, err := keyCollection.Aggregate(ctx, refreshQuotaPipeline)
	if err != nil {
		// @TODO: Log Error
		panic(err)
	}
}
