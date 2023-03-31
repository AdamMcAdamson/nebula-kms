package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service represents data about a KMS Service
type Service struct {
	ID                primitive.ObjectID `json:"_id" bson:"_id"`
	Name              string             `json:"service_name" bson:"service_name"`
	Type              string             `json:"service_type" bson:"service_type"` // @TODO: Enum (?) (PublicProduction, PrivateProduction, Staging)
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
	SourceIdentifiers []string           `json:"source_identifiers" bson:"source_identifiers"`
}
