package model

import (
	"context"
	"time"

	"jan-server/services/llm-api/internal/domain/query"

	decimal "github.com/shopspring/decimal"
)

type SupportedParameters struct {
	Names   []string                    `json:"names"`   // e.g., ["include_reasoning","max_tokens",...]
	Default map[string]*decimal.Decimal `json:"default"` // temperature/top_p/frequency_penalty, null allowed
}

// Architecture metadata.
type Architecture struct {
	Modality         string   `json:"modality"` // "text+image->text"
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
	Tokenizer        string   `json:"tokenizer"`     // "GPT" / "SentencePiece" / etc.
	InstructType     *string  `json:"instruct_type"` // nullable
}

type ModelCatalogStatus string

const (
	ModelCatalogStatusInit    ModelCatalogStatus = "init"    // default status when creating entry
	ModelCatalogStatusFilled  ModelCatalogStatus = "filled"  // may update from Provider like OpenRouter
	ModelCatalogStatusUpdated ModelCatalogStatus = "updated" // manually updated by admin (cannot be auto-updated anymore
)

type ModelCatalog struct {
	ID                  uint                `json:"id"`
	PublicID            string              `json:"public_id"`
	SupportedParameters SupportedParameters `json:"supported_parameters"`
	Architecture        Architecture        `json:"architecture"`
	Tags                []string            `json:"tags,omitempty"`
	Notes               *string             `json:"notes,omitempty"`
	IsModerated         *bool               `json:"is_moderated,omitempty"`
	Active              *bool               `json:"active,omitempty"`
	Extras              map[string]any      `json:"extras,omitempty"`
	Status              ModelCatalogStatus  `json:"status"`
	Experimental        bool                `json:"experimental"`
	// Capabilities (moved from provider_model)
	SupportsImages     bool   `json:"supports_images"`
	SupportsEmbeddings bool   `json:"supports_embeddings"`
	SupportsReasoning  bool   `json:"supports_reasoning"`
	SupportsAudio      bool   `json:"supports_audio"`
	SupportsVideo      bool   `json:"supports_video"`
	Family             string `json:"family,omitempty"` // e.g., "gpt-4o", "llama-3.1"
	LastSyncedAt       *time.Time
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ModelCatalogFilter struct {
	IDs                *[]uint
	PublicID           *string
	IsModerated        *bool
	Active             *bool
	Status             *ModelCatalogStatus
	LastSyncedAfter    *time.Time
	LastSyncedBefore   *time.Time
	Experimental       *bool
	SupportsImages     *bool
	SupportsEmbeddings *bool
	SupportsReasoning  *bool
	SupportsAudio      *bool
	SupportsVideo      *bool
	Family             *string
}

type ModelCatalogRepository interface {
	Create(ctx context.Context, catalog *ModelCatalog) error
	Update(ctx context.Context, catalog *ModelCatalog) error
	DeleteByID(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*ModelCatalog, error)
	FindByPublicID(ctx context.Context, publicID string) (*ModelCatalog, error)
	FindByFilter(ctx context.Context, filter ModelCatalogFilter, p *query.Pagination) ([]*ModelCatalog, error)
	Count(ctx context.Context, filter ModelCatalogFilter) (int64, error)
	BatchUpdateActive(ctx context.Context, filter ModelCatalogFilter, active bool) (int64, error)
	// Batch methods for optimization
	FindByIDs(ctx context.Context, ids []uint) ([]*ModelCatalog, error)
	FindByPublicIDs(ctx context.Context, publicIDs []string) ([]*ModelCatalog, error)
}
