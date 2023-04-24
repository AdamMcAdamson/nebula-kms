package models

import (
	"encoding/json"
	"time"

	"github.com/UTDNebula/kms/configs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service represents data about a KMS Service
type Service struct {
	ID                primitive.ObjectID `json:"_id" bson:"_id"`
	Name              string             `json:"service_name" bson:"service_name"`
	Type              string             `json:"service_type" bson:"service_type"` // @TODO: Enum (?) (Basic, PublicProduction, PrivateProduction, Staging)
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
	SourceIdentifiers []string           `json:"source_identifiers" bson:"source_identifiers"`
}

func (s Service) MarshalJSON() ([]byte, error) {
	type Alias Service
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Alias
	}{
		// use the desired date layout
		CreatedAt: s.CreatedAt.Format(configs.DateLayout),
		UpdatedAt: s.UpdatedAt.Format(configs.DateLayout),
		Alias:     Alias(s),
	})
}
