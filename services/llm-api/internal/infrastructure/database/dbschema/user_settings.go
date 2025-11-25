package dbschema

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"jan-server/services/llm-api/internal/domain/usersettings"
	"jan-server/services/llm-api/internal/infrastructure/database"
)

func init() {
	database.RegisterSchemaForAutoMigrate(UserSettings{})
}

// UserSettings is the database schema for user_settings table.
type UserSettings struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"not null;uniqueIndex:ux_user_settings_user_id"`

	// Memory Feature Controls
	MemoryEnabled            bool `gorm:"not null;default:true"`
	MemoryAutoInject         bool `gorm:"not null;default:false"`
	MemoryInjectUserCore     bool `gorm:"not null;default:false"`
	MemoryInjectProject      bool `gorm:"not null;default:false"`
	MemoryInjectConversation bool `gorm:"not null;default:false"`

	// Memory Retrieval Preferences
	MemoryMaxUserItems     int     `gorm:"not null;default:3"`
	MemoryMaxProjectItems  int     `gorm:"not null;default:5"`
	MemoryMaxEpisodicItems int     `gorm:"not null;default:3"`
	MemoryMinSimilarity    float32 `gorm:"type:numeric(3,2);not null;default:0.75"`

	// Other Feature Toggles
	EnableTrace bool `gorm:"not null;default:false"`
	EnableTools bool `gorm:"not null;default:true"`

	// Preferences - flexible JSON
	Preferences JSONB `gorm:"type:jsonb;not null;default:'{}'"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for UserSettings.
func (UserSettings) TableName() string {
	return "llm_api.user_settings"
}

// JSONB is a custom type for JSONB columns.
type JSONB map[string]interface{}

// Value implements driver.Valuer interface for JSONB.
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface for JSONB.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// EtoD converts entity (database schema) to domain model.
func (e *UserSettings) EtoD() *usersettings.UserSettings {
	return &usersettings.UserSettings{
		ID:                       e.ID,
		UserID:                   e.UserID,
		MemoryEnabled:            e.MemoryEnabled,
		MemoryAutoInject:         e.MemoryAutoInject,
		MemoryInjectUserCore:     e.MemoryInjectUserCore,
		MemoryInjectProject:      e.MemoryInjectProject,
		MemoryInjectConversation: e.MemoryInjectConversation,
		MemoryMaxUserItems:       e.MemoryMaxUserItems,
		MemoryMaxProjectItems:    e.MemoryMaxProjectItems,
		MemoryMaxEpisodicItems:   e.MemoryMaxEpisodicItems,
		MemoryMinSimilarity:      e.MemoryMinSimilarity,
		EnableTrace:              e.EnableTrace,
		EnableTools:              e.EnableTools,
		Preferences:              map[string]interface{}(e.Preferences),
		CreatedAt:                e.CreatedAt,
		UpdatedAt:                e.UpdatedAt,
	}
}

// NewSchemaUserSettings converts domain model to entity (database schema).
func NewSchemaUserSettings(d *usersettings.UserSettings) *UserSettings {
	prefs := JSONB(d.Preferences)
	if prefs == nil {
		prefs = make(JSONB)
	}

	return &UserSettings{
		ID:                       d.ID,
		UserID:                   d.UserID,
		MemoryEnabled:            d.MemoryEnabled,
		MemoryAutoInject:         d.MemoryAutoInject,
		MemoryInjectUserCore:     d.MemoryInjectUserCore,
		MemoryInjectProject:      d.MemoryInjectProject,
		MemoryInjectConversation: d.MemoryInjectConversation,
		MemoryMaxUserItems:       d.MemoryMaxUserItems,
		MemoryMaxProjectItems:    d.MemoryMaxProjectItems,
		MemoryMaxEpisodicItems:   d.MemoryMaxEpisodicItems,
		MemoryMinSimilarity:      d.MemoryMinSimilarity,
		EnableTrace:              d.EnableTrace,
		EnableTools:              d.EnableTools,
		Preferences:              prefs,
		CreatedAt:                d.CreatedAt,
		UpdatedAt:                d.UpdatedAt,
	}
}
