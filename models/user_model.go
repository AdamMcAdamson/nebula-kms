package models

import (
	"encoding/json"
	"time"

	"github.com/UTDNebula/kms/configs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents data about a KMS User
type User struct {
	ID             primitive.ObjectID   `json:"_id" bson:"_id"`
	PlatformUserID primitive.ObjectID   `json:"platform_user_id" bson:"platform_user_id"`
	Type           string               `json:"user_type" bson:"user_type"` // @TODO: Enum (?) (Developer, Lead, Admin)
	CreatedAt      time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at" bson:"updated_at"`
	BasicKey       primitive.ObjectID   `json:"basic_key" bson:"basic_key"`
	AdvancedKeys   []primitive.ObjectID `json:"advanced_keys" bson:"advanced_keys"`
	Services       []primitive.ObjectID `json:"services,omitempty" bson:"services,omitempty"` // @TODO: Verify best solution. Services should only appear if you are a Lead.
}

func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Alias
	}{
		// use the desired date layout
		CreatedAt: u.CreatedAt.Format(configs.DateLayout),
		UpdatedAt: u.UpdatedAt.Format(configs.DateLayout),
		Alias:     Alias(u),
	})
}
