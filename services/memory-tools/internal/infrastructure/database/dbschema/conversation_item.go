package dbschema

import (
	"time"

	"github.com/janhq/jan-server/services/memory-tools/internal/domain/memory"
)

type ConversationItem struct {
	ID             string    `db:"id"`
	ConversationID string    `db:"conversation_id"`
	Role           string    `db:"role"`
	Content        string    `db:"content"`
	ToolCalls      string    `db:"tool_calls"`
	CreatedAt      time.Time `db:"created_at"`
}

func NewSchemaConversationItem(d *memory.ConversationItem) *ConversationItem {
	if d == nil {
		return nil
	}

	return &ConversationItem{
		ID:             d.ID,
		ConversationID: d.ConversationID,
		Role:           d.Role,
		Content:        d.Content,
		ToolCalls:      d.ToolCalls,
		CreatedAt:      d.CreatedAt,
	}
}

func (s *ConversationItem) EtoD() *memory.ConversationItem {
	if s == nil {
		return nil
	}

	return &memory.ConversationItem{
		ID:             s.ID,
		ConversationID: s.ConversationID,
		Role:           s.Role,
		Content:        s.Content,
		ToolCalls:      s.ToolCalls,
		CreatedAt:      s.CreatedAt,
	}
}
