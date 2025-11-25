// Package usersettings provides domain models for user preferences and settings.
package usersettings

import (
	"context"
	"time"
)

// UserSettings represents user preferences and feature toggles.
type UserSettings struct {
	ID     uint
	UserID uint

	// Memory Feature Controls
	MemoryEnabled            bool
	MemoryAutoInject         bool
	MemoryInjectUserCore     bool
	MemoryInjectProject      bool
	MemoryInjectConversation bool

	// Memory Retrieval Preferences
	MemoryMaxUserItems     int
	MemoryMaxProjectItems  int
	MemoryMaxEpisodicItems int
	MemoryMinSimilarity    float32

	// Other Feature Toggles
	EnableTrace bool
	EnableTools bool

	// Preferences - flexible JSON for future extensions
	Preferences map[string]interface{}

	CreatedAt time.Time
	UpdatedAt time.Time
}

// MemoryConfig returns memory configuration derived from settings.
type MemoryConfig struct {
	Enabled            bool
	AutoInject         bool
	InjectUserCore     bool
	InjectProject      bool
	InjectConversation bool
	MaxUserItems       int
	MaxProjectItems    int
	MaxEpisodicItems   int
	MinSimilarity      float32
}

// GetMemoryConfig extracts memory configuration from user settings.
func (s *UserSettings) GetMemoryConfig() MemoryConfig {
	return MemoryConfig{
		Enabled:            s.MemoryEnabled,
		AutoInject:         s.MemoryAutoInject,
		InjectUserCore:     s.MemoryInjectUserCore,
		InjectProject:      s.MemoryInjectProject,
		InjectConversation: s.MemoryInjectConversation,
		MaxUserItems:       s.MemoryMaxUserItems,
		MaxProjectItems:    s.MemoryMaxProjectItems,
		MaxEpisodicItems:   s.MemoryMaxEpisodicItems,
		MinSimilarity:      s.MemoryMinSimilarity,
	}
}

// DefaultUserSettings returns settings with safe defaults (memory disabled by default).
func DefaultUserSettings(userID uint) *UserSettings {
	return &UserSettings{
		UserID:                   userID,
		MemoryEnabled:            true,
		MemoryAutoInject:         false, // Default OFF per improvement plan
		MemoryInjectUserCore:     false,
		MemoryInjectProject:      false,
		MemoryInjectConversation: false,
		MemoryMaxUserItems:       3,
		MemoryMaxProjectItems:    5,
		MemoryMaxEpisodicItems:   3,
		MemoryMinSimilarity:      0.75,
		EnableTrace:              false,
		EnableTools:              true,
		Preferences:              make(map[string]interface{}),
	}
}

// UpdateRequest represents fields that can be updated via API.
type UpdateRequest struct {
	MemoryEnabled            *bool                  `json:"memory_enabled,omitempty"`
	MemoryAutoInject         *bool                  `json:"memory_auto_inject,omitempty"`
	MemoryInjectUserCore     *bool                  `json:"memory_inject_user_core,omitempty"`
	MemoryInjectProject      *bool                  `json:"memory_inject_project,omitempty"`
	MemoryInjectConversation *bool                  `json:"memory_inject_conversation,omitempty"`
	MemoryMaxUserItems       *int                   `json:"memory_max_user_items,omitempty"`
	MemoryMaxProjectItems    *int                   `json:"memory_max_project_items,omitempty"`
	MemoryMaxEpisodicItems   *int                   `json:"memory_max_episodic_items,omitempty"`
	MemoryMinSimilarity      *float32               `json:"memory_min_similarity,omitempty"`
	EnableTrace              *bool                  `json:"enable_trace,omitempty"`
	EnableTools              *bool                  `json:"enable_tools,omitempty"`
	Preferences              map[string]interface{} `json:"preferences,omitempty"`
}

// Apply updates the UserSettings with non-nil fields from UpdateRequest.
func (s *UserSettings) Apply(req UpdateRequest) {
	if req.MemoryEnabled != nil {
		s.MemoryEnabled = *req.MemoryEnabled
	}
	if req.MemoryAutoInject != nil {
		s.MemoryAutoInject = *req.MemoryAutoInject
	}
	if req.MemoryInjectUserCore != nil {
		s.MemoryInjectUserCore = *req.MemoryInjectUserCore
	}
	if req.MemoryInjectProject != nil {
		s.MemoryInjectProject = *req.MemoryInjectProject
	}
	if req.MemoryInjectConversation != nil {
		s.MemoryInjectConversation = *req.MemoryInjectConversation
	}
	if req.MemoryMaxUserItems != nil {
		s.MemoryMaxUserItems = *req.MemoryMaxUserItems
	}
	if req.MemoryMaxProjectItems != nil {
		s.MemoryMaxProjectItems = *req.MemoryMaxProjectItems
	}
	if req.MemoryMaxEpisodicItems != nil {
		s.MemoryMaxEpisodicItems = *req.MemoryMaxEpisodicItems
	}
	if req.MemoryMinSimilarity != nil {
		s.MemoryMinSimilarity = *req.MemoryMinSimilarity
	}
	if req.EnableTrace != nil {
		s.EnableTrace = *req.EnableTrace
	}
	if req.EnableTools != nil {
		s.EnableTools = *req.EnableTools
	}
	if req.Preferences != nil {
		s.Preferences = req.Preferences
	}
}

// Repository defines storage operations for user settings.
type Repository interface {
	FindByUserID(ctx context.Context, userID uint) (*UserSettings, error)
	Upsert(ctx context.Context, settings *UserSettings) (*UserSettings, error)
	Update(ctx context.Context, settings *UserSettings) error
}

// Service manages user settings operations.
type Service struct {
	repo Repository
}

// NewService constructs a Service with required dependencies.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetOrCreateSettings retrieves existing settings or creates defaults for a user.
func (s *Service) GetOrCreateSettings(ctx context.Context, userID uint) (*UserSettings, error) {
	settings, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create default settings if none exist
	if settings == nil {
		defaults := DefaultUserSettings(userID)
		return s.repo.Upsert(ctx, defaults)
	}

	return settings, nil
}

// UpdateSettings applies updates to user settings.
func (s *Service) UpdateSettings(ctx context.Context, userID uint, req UpdateRequest) (*UserSettings, error) {
	settings, err := s.GetOrCreateSettings(ctx, userID)
	if err != nil {
		return nil, err
	}

	settings.Apply(req)

	if err := s.repo.Update(ctx, settings); err != nil {
		return nil, err
	}

	return settings, nil
}
