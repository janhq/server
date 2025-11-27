package conversation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	domain "jan-server/services/response-api/internal/domain/conversation"
	"jan-server/services/response-api/internal/infrastructure/database/entities"
	"jan-server/services/response-api/internal/utils/platformerrors"
)

// Repository persists conversation metadata.
type Repository struct {
	db *gorm.DB
}

// NewRepository builds a conversation repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts the conversation record.
func (r *Repository) Create(ctx context.Context, conv *domain.Conversation) error {
	metadata, err := marshalJSON(conv.Metadata)
	if err != nil {
		return platformerrors.NewError(
			ctx,
			platformerrors.LayerRepository,
			platformerrors.ErrorTypeInternal,
			"failed to marshal conversation metadata",
			err,
			"1f2e3d4c-5a6b-7c8d-9e0f-1a2b3c4d5e6f",
		)
	}

	entity := &entities.Conversation{
		PublicID: conv.PublicID,
		UserID:   conv.UserID,
		Metadata: metadata,
	}

	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return platformerrors.NewError(
			ctx,
			platformerrors.LayerRepository,
			platformerrors.ErrorTypeDatabaseError,
			"failed to create conversation",
			err,
			"2g3f4e5d-6b7c-8d9e-0f1a-2b3c4d5e6f7a",
		)
	}

	conv.ID = entity.ID
	conv.CreatedAt = entity.CreatedAt
	conv.UpdatedAt = entity.UpdatedAt
	return nil
}

// FindByPublicID fetches a conversation.
func (r *Repository) FindByPublicID(ctx context.Context, publicID string) (*domain.Conversation, error) {
	var entity entities.Conversation
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, platformerrors.NewError(
				ctx,
				platformerrors.LayerRepository,
				platformerrors.ErrorTypeNotFound,
				fmt.Sprintf("conversation not found: %s", publicID),
				nil,
				"3h4g5f6e-7c8d-9e0f-1a2b-3c4d5e6f7a8b",
			)
		}
		return nil, platformerrors.NewError(
			ctx,
			platformerrors.LayerRepository,
			platformerrors.ErrorTypeDatabaseError,
			"failed to fetch conversation",
			err,
			"4i5h6g7f-8d9e-0f1a-2b3c-4d5e6f7a8b9c",
		)
	}

	var metadata map[string]interface{}
	if len(entity.Metadata) > 0 {
		if err := json.Unmarshal(entity.Metadata, &metadata); err != nil {
			return nil, platformerrors.NewError(
				ctx,
				platformerrors.LayerRepository,
				platformerrors.ErrorTypeInternal,
				"failed to unmarshal conversation metadata",
				err,
				"5j6i7h8g-9e0f-1a2b-3c4d-5e6f7a8b9c0d",
			)
		}
	}

	return &domain.Conversation{
		ID:        entity.ID,
		PublicID:  entity.PublicID,
		UserID:    entity.UserID,
		Metadata:  metadata,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}, nil
}
