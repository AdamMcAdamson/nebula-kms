package models

import (
	"encoding/json"
	"time"

	"github.com/UTDNebula/kms/configs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Key represents data about a KMS Key
type Key struct {
	ID      primitive.ObjectID `json:"_id" bson:"_id"`
	Key     string             `json:"key" bson:"key"`
	Type    string             `json:"key_type" bson:"key_type"` // @TODO: Enum (?) (Basic, Advanced)
	Name    string             `json:"name" bson:"name"`
	OwnerID primitive.ObjectID `json:"owner_id" bson:"owner_id"`

	// @TODO: Determine if we want to use different models for Basic and Advanced keys (basic keys not containing a serviceID)
	ServiceID primitive.ObjectID `json:"service_id,omitempty" bson:"service_id,omitempty"`

	Quota          int       `json:"quota" bson:"quota"`
	QuotaNumDays   int       `json:"quota_num_days" bson:"quota_num_days"`
	UsageRemaining int       `json:"usage_remaining" bson:"usage_remaining"`
	QuotaTimestamp time.Time `json:"quota_timestamp" bson:"quota_timestamp"`
	LastUsed       time.Time `json:"last_used" bson:"last_used"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" bson:"updated_at"`
	IsActive       bool      `json:"is_active" bson:"is_active"`
}

func (k Key) MarshalJSON() ([]byte, error) {
	type Alias Key
	return json.Marshal(&struct {
		QuotaTimestamp string `json:"quota_timestamp"`
		LastUsed       string `json:"last_used"`
		CreatedAt      string `json:"created_at"`
		UpdatedAt      string `json:"updated_at"`
		Alias
	}{
		// use the desired date layout
		QuotaTimestamp: k.QuotaTimestamp.Format(configs.DateLayout),
		LastUsed:       k.LastUsed.Format(configs.DateLayout),
		CreatedAt:      k.CreatedAt.Format(configs.DateLayout),
		UpdatedAt:      k.UpdatedAt.Format(configs.DateLayout),
		Alias:          Alias(k),
	})
}
