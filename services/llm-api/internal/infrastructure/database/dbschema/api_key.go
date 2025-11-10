package dbschema

import (
	"time"

	"jan-server/services/llm-api/internal/domain/apikey"
	"jan-server/services/llm-api/internal/infrastructure/database"
)

func init() {
	database.RegisterSchemaForAutoMigrate(&APIKey{})
}

// APIKey represents persisted API key metadata.
type APIKey struct {
	ID               string    `gorm:"type:uuid;primaryKey"`
	UserID           uint      `gorm:"not null;index"`
	Name             string    `gorm:"type:varchar(128);not null"`
	Prefix           string    `gorm:"type:varchar(32);not null"`
	Suffix           string    `gorm:"type:varchar(8);not null"`
	Hash             string    `gorm:"type:varchar(128);not null"`
	KongCredentialID string    `gorm:"type:varchar(64);not null"`
	ExpiresAt        time.Time `gorm:"not null;index"`
	RevokedAt        *time.Time
	LastUsedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// EtoD converts schema model to domain representation.
func (k *APIKey) EtoD() *apikey.APIKey {
	if k == nil {
		return nil
	}
	return &apikey.APIKey{
		ID:               k.ID,
		UserID:           k.UserID,
		Name:             k.Name,
		Prefix:           k.Prefix,
		Suffix:           k.Suffix,
		Hash:             k.Hash,
		KongCredentialID: k.KongCredentialID,
		ExpiresAt:        k.ExpiresAt,
		RevokedAt:        k.RevokedAt,
		LastUsedAt:       k.LastUsedAt,
		CreatedAt:        k.CreatedAt,
		UpdatedAt:        k.UpdatedAt,
	}
}

// FromDomain converts domain model to schema representation.
func FromDomain(apiKey *apikey.APIKey) *APIKey {
	if apiKey == nil {
		return nil
	}
	return &APIKey{
		ID:               apiKey.ID,
		UserID:           apiKey.UserID,
		Name:             apiKey.Name,
		Prefix:           apiKey.Prefix,
		Suffix:           apiKey.Suffix,
		Hash:             apiKey.Hash,
		KongCredentialID: apiKey.KongCredentialID,
		ExpiresAt:        apiKey.ExpiresAt,
		RevokedAt:        apiKey.RevokedAt,
		LastUsedAt:       apiKey.LastUsedAt,
		CreatedAt:        apiKey.CreatedAt,
		UpdatedAt:        apiKey.UpdatedAt,
	}
}
