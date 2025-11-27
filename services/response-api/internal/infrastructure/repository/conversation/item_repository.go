package conversation

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	domain "jan-server/services/response-api/internal/domain/conversation"
	"jan-server/services/response-api/internal/infrastructure/database/entities"
	"jan-server/services/response-api/internal/utils/platformerrors"
)

// ItemRepository persists conversation items.
type ItemRepository struct {
	db *gorm.DB
}

// NewItemRepository constructs the item repository.
func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// BulkInsert stores multiple conversation items in sequence order.
func (r *ItemRepository) BulkInsert(ctx context.Context, items []domain.Item) error {
	if len(items) == 0 {
		return nil
	}

	rows := make([]entities.ConversationItem, 0, len(items))
	for _, item := range items {
		content, err := marshalJSON(item.Content)
		if err != nil {
			return platformerrors.NewError(
				ctx,
				platformerrors.LayerRepository,
				platformerrors.ErrorTypeInternal,
				"failed to marshal conversation item",
				err,
				"6k7j8i9h-0f1a-2b3c-4d5e-6f7a8b9c0d1e",
			)
		}
		rows = append(rows, entities.ConversationItem{
			ConversationID: item.ConversationID,
			Role:           string(item.Role),
			Status:         string(item.Status),
			Content:        content,
			Sequence:       item.Sequence,
			CreatedAt:      item.CreatedAt,
		})
	}

	if err := r.db.WithContext(ctx).Create(&rows).Error; err != nil {
		return platformerrors.NewError(
			ctx,
			platformerrors.LayerRepository,
			platformerrors.ErrorTypeDatabaseError,
			"failed to bulk insert conversation items",
			err,
			"7l8k9j0i-1a2b-3c4d-5e6f-7a8b9c0d1e2f",
		)
	}
	return nil
}

// ListByConversationID returns items ordered by sequence.
func (r *ItemRepository) ListByConversationID(ctx context.Context, conversationID uint) ([]domain.Item, error) {
	var rows []entities.ConversationItem
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("sequence ASC").
		Find(&rows).Error; err != nil {
		return nil, platformerrors.NewError(
			ctx,
			platformerrors.LayerRepository,
			platformerrors.ErrorTypeDatabaseError,
			"failed to list conversation items",
			err,
			"8m9l0k1j-2b3c-4d5e-6f7a-8b9c0d1e2f3a",
		)
	}

	items := make([]domain.Item, 0, len(rows))
	for _, row := range rows {
		var content map[string]interface{}
		if len(row.Content) > 0 {
			_ = json.Unmarshal(row.Content, &content)
		}
		items = append(items, domain.Item{
			ID:             row.ID,
			ConversationID: row.ConversationID,
			Role:           domain.ItemRole(row.Role),
			Status:         domain.ItemStatus(row.Status),
			Content:        content,
			Sequence:       row.Sequence,
			CreatedAt:      row.CreatedAt,
		})
	}
	return items, nil
}
