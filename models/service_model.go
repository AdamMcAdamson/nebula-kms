package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Service represents data about a KMS Service
type Service struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	Name             string             `json:"service_name" bson:"service_name"`
	Type             string             `json:"service_type" bson:"service_type"` // @TODO: Enum (?) (PublicProduction, PrivateProduction, Staging)
	SourceIdentifier []string           `json:"source_identifier" bson:"source_identifier"`
}
