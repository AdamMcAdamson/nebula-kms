package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Key represents data about a KMS Key
type Key struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	Key            string             `json:"key" bson:"key"`
	Type           string             `json:"key_type" bson:"key_type"` // @TODO: Enum (?) (Basic, Advanced)
	Name           string             `json:"name" bson:"name"`
	OwnerID        primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	Quota          int                `json:"quota" bson:"quota"`
	QuotaType      string             `json:"quota_type" bson:"quota_type"` // @TODO: Enum (?) (Daily, etc...)
	UsageRemaining int                `json:"usage_remaining" bson:"usage_remaining"`
	QuotaTimestamp time.Time          `json:"quota_timestamp" bson:"quota_timestamp"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
	IsActive       bool               `json:"is_active" bson:"is_active"`
}
