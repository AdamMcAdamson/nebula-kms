package models

import (
	"encoding/json"
	"time"

	"github.com/UTDNebula/kms/configs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Key represents data about a KMS Key
// @TODO: Force quota_timestamp to the end of the current period UTC-0
//  - i.e. 1 Day - Midnight tonight, Weekly - Midnight Sunday, etc..
// @TODO: Run refresh usage_remaining on server startup and one Time passing Midnight UTC-0

type Key struct {
	ID      primitive.ObjectID `json:"_id" bson:"_id"`
	Key     string             `json:"key" bson:"key"`
	Type    string             `json:"key_type" bson:"key_type"` // @TODO: Enum (?) (Basic, Advanced)
	Name    string             `json:"name" bson:"name"`
	OwnerID primitive.ObjectID `json:"owner_id" bson:"owner_id"`

	// @TODO: ServiceIDs needs to be an array to handle basic keys (since they can be used for either the nebula api or the platform api).
	// ServiceIDs []primitive.ObjectID `json:"service_ids" bson:"service_ids"`

	// @TODO: A more elegant solution would be to have accounts have a basic key for each type of service (i.e. Basic_Nebula_API_Key and Basic_Platform_API_Key).
	// 			Or we can update service_type to include a 'Basic' type, which basic keys can be assumed to be associated with.
	//			With this solution, using different models for Basic and Advanced keys would be advised
	ServiceID primitive.ObjectID `json:"service_id,omitempty" bson:"service_id,omitempty"`

	Quota          int       `json:"quota" bson:"quota"`
	QuotaNumDays   int       `json:"quota_num_days" bson:"quota_num_days"` // @TODO: Enum (?) (Daily, etc...)
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
