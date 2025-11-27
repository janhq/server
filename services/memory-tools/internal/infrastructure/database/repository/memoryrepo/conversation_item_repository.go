package memoryrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/janhq/jan-server/services/memory-tools/internal/domain/memory"
	"github.com/janhq/jan-server/services/memory-tools/internal/infrastructure/database/dbschema"
)

func (r *Repository) CreateConversationItem(ctx context.Context, item *memory.ConversationItem) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	schema := dbschema.NewSchemaConversationItem(item)

	if err := r.db.WithContext(ctx).
		Table("conversation_items").
		Create(map[string]any{
			"id":              schema.ID,
			"conversation_id": schema.ConversationID,
			"role":            schema.Role,
			"content":         schema.Content,
			"tool_calls":      schema.ToolCalls,
			"created_at":      schema.CreatedAt,
		}).Error; err != nil {
		return fmt.Errorf("create conversation item: %w", err)
	}

	return nil
}

func (r *Repository) GetConversationItems(ctx context.Context, conversationID string) ([]memory.ConversationItem, error) {
	var rows []dbschema.ConversationItem
	if err := r.db.WithContext(ctx).
		Table("conversation_items").
		Select("id, conversation_id, role, content, tool_calls, created_at").
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query conversation items: %w", err)
	}

	items := make([]memory.ConversationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, *row.EtoD())
	}

	return items, nil
}
