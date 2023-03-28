package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents data about a KMS User
type User struct {
	ID             primitive.ObjectID   `json:"_id" bson:"_id"`
	PlatformUserID primitive.ObjectID   `json:"platform_user_id" bson:"platform_user_id"`
	Type           string               `json:"user_type" bson:"user_type"` // @TODO: Enum (?) (Developer, Lead, Admin)
	CreatedAt      time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at" bson:"updated_at"`
	Keys           []primitive.ObjectID `json:"keys" bson:"keys"`
	Services       []primitive.ObjectID `json:"services,omitempty" bson:"services,omitempty"` // @TODO: Verify best solution. Services should only appear if you are a Lead.
}
