package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Key represents data about a KMS Key
type Key struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	Key            string             `json:"key" bson:"key"`
	Type           string             `json:"key_type" bson:"key_type"` // @TODO: Enum (?) (Basic, Advanced)
	Name           string             `json:"name" bson:"name"`
	Quota          int                `json:"quota" bson:"quota"`
	QuotaType      string             `json:"quota_type" bson:"quota_type"` // @TODO: Enum (?) (Daily, etc...)
	UsageRemaining int                `json:"usage_remaining" bson:"usage_remaining"`
	QuotaTimestamp primitive.DateTime `json:"quota_timestamp" bson:"quota_timestamp"`
	CreatedAt      primitive.DateTime `json:"created_at" bson:"created_at"`
	LastModified   primitive.DateTime `json:"last_modified" bson:"last_modified"`
	IsActive       bool               `json:"is_active" bson:"is_active"`
}
