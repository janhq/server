package conversationrepo

import (
	"context"

	"jan-server/services/llm-api/internal/domain/conversation"
	"jan-server/services/llm-api/internal/domain/query"
	"jan-server/services/llm-api/internal/infrastructure/database/dbschema"
	"jan-server/services/llm-api/internal/infrastructure/database/gormgen"
	"jan-server/services/llm-api/internal/infrastructure/database/transaction"
	"jan-server/services/llm-api/internal/utils/functional"
	"jan-server/services/llm-api/internal/utils/platformerrors"
)

type ConversationGormRepository struct {
	db *transaction.Database
}

var _ conversation.ConversationRepository = (*ConversationGormRepository)(nil)

func NewConversationGormRepository(db *transaction.Database) conversation.ConversationRepository {
	return &ConversationGormRepository{db}
}

// Create implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) Create(ctx context.Context, conv *conversation.Conversation) error {
	model := dbschema.NewSchemaConversation(conv)
	if err := repo.db.GetQuery(ctx).Conversation.WithContext(ctx).Create(model); err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to create conversation")
	}
	// Update the domain object with generated ID and timestamps
	conv.ID = model.ID
	conv.CreatedAt = model.CreatedAt
	conv.UpdatedAt = model.UpdatedAt
	return nil
}

// FindByFilter implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) FindByFilter(ctx context.Context, filter conversation.ConversationFilter, pagination *query.Pagination) ([]*conversation.Conversation, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.Conversation.WithContext(ctx)
	sql = repo.applyFilter(q, sql, filter)
	sql = repo.applyPagination(q, sql, pagination)

	rows, err := sql.Find()
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find conversations")
	}

	result := functional.Map(rows, func(item *dbschema.Conversation) *conversation.Conversation {
		return item.EtoD()
	})
	return result, nil
}

// Count implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) Count(ctx context.Context, filter conversation.ConversationFilter) (int64, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.Conversation.WithContext(ctx)
	sql = repo.applyFilter(q, sql, filter)
	return sql.Count()
}

// FindByID implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) FindByID(ctx context.Context, id uint) (*conversation.Conversation, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.Conversation.WithContext(ctx)
	sql = repo.applyFilter(q, sql, conversation.ConversationFilter{ID: &id})
	result, err := sql.First()
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find conversation by ID")
	}
	return result.EtoD(), nil
}

// FindByPublicID implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) FindByPublicID(ctx context.Context, publicID string) (*conversation.Conversation, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.Conversation.WithContext(ctx)
	sql = repo.applyFilter(q, sql, conversation.ConversationFilter{PublicID: &publicID})
	result, err := sql.First()
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find conversation by public ID")
	}
	return result.EtoD(), nil
}

// Update implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) Update(ctx context.Context, conv *conversation.Conversation) error {
	model := dbschema.NewSchemaConversation(conv)
	q := repo.db.GetQuery(ctx)

	// Use Save to update all fields
	if err := q.Conversation.WithContext(ctx).Where(q.Conversation.ID.Eq(conv.ID)).Save(model); err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to update conversation")
	}

	// Update timestamps
	conv.UpdatedAt = model.UpdatedAt
	return nil
}

// Delete implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) Delete(ctx context.Context, id uint) error {
	q := repo.db.GetQuery(ctx)
	_, err := q.Conversation.WithContext(ctx).Where(q.Conversation.ID.Eq(id)).Delete()
	if err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to delete conversation")
	}
	return nil
}

// AddItem implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) AddItem(ctx context.Context, conversationID uint, item *conversation.Item) error {
	// Verify conversation exists
	_, err := repo.FindByID(ctx, conversationID)
	if err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "conversation not found")
	}

	// Set conversation ID
	item.ConversationID = conversationID

	// Create the item
	model := dbschema.NewSchemaConversationItem(item)
	q := repo.db.GetQuery(ctx)

	if err := q.ConversationItem.WithContext(ctx).Create(model); err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to create conversation item")
	}

	// Update the domain object with generated ID
	item.ID = model.ID
	item.CreatedAt = model.CreatedAt

	return nil
}

// SearchItems implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) SearchItems(ctx context.Context, conversationID uint, searchQuery string) ([]*conversation.Item, error) {
	// For now, this is a simple implementation
	// In production, you'd want to use full-text search or a search engine like Elasticsearch
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyItemFilter(q, sql, conversation.ItemFilter{
		ConversationID: &conversationID,
	})

	// Search in content JSON field (PostgreSQL JSONB search)
	// This is a basic implementation - enhance based on your database capabilities
	// Note: For proper JSON search in PostgreSQL, you might need raw SQL or custom query
	rows, err := sql.Find()

	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to search items")
	}

	result := functional.Map(rows, func(item *dbschema.ConversationItem) *conversation.Item {
		return item.EtoD()
	})

	// TODO: Implement proper full-text search filtering
	// For now, returning all items in the conversation
	return result, nil
}

// BulkAddItems implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) BulkAddItems(ctx context.Context, conversationID uint, items []*conversation.Item) error {
	if len(items) == 0 {
		return nil
	}

	// Verify conversation exists
	_, err := repo.FindByID(ctx, conversationID)
	if err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "conversation not found")
	}

	// Set conversation ID for all items
	for _, item := range items {
		item.ConversationID = conversationID
	}

	// Convert to schema models
	models := functional.Map(items, func(item *conversation.Item) *dbschema.ConversationItem {
		return dbschema.NewSchemaConversationItem(item)
	})

	// Bulk insert with manual batching to ensure ID population
	q := repo.db.GetQuery(ctx)
	batchSize := 100

	// Process in batches
	for i := 0; i < len(models); i += batchSize {
		end := i + batchSize
		if end > len(models) {
			end = len(models)
		}

		batch := models[i:end]
		if err := q.ConversationItem.WithContext(ctx).Create(batch...); err != nil {
			return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to bulk create items")
		}

		// Update domain objects with generated IDs for this batch
		for j, model := range batch {
			items[i+j].ID = model.ID
			items[i+j].CreatedAt = model.CreatedAt
		}
	}

	return nil
}

// GetItemByID implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) GetItemByID(ctx context.Context, conversationID uint, itemID uint) (*conversation.Item, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyItemFilter(q, sql, conversation.ItemFilter{
		ID:             &itemID,
		ConversationID: &conversationID,
	})
	result, err := sql.First()
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find item by ID")
	}
	return result.EtoD(), nil
}

// GetItemByPublicID implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) GetItemByPublicID(ctx context.Context, conversationID uint, publicID string) (*conversation.Item, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyItemFilter(q, sql, conversation.ItemFilter{
		PublicID:       &publicID,
		ConversationID: &conversationID,
	})
	result, err := sql.First()
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find item by public ID")
	}
	return result.EtoD(), nil
}

// GetItemByCallID implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) GetItemByCallID(ctx context.Context, conversationID uint, callID string) (*conversation.Item, error) {
	q := repo.db.GetQuery(ctx)
	// Use raw SQL since gormgen may not have the call_id field
	var result dbschema.ConversationItem
	err := q.ConversationItem.WithContext(ctx).UnderlyingDB().
		Where("conversation_id = ? AND call_id = ?", conversationID, callID).
		First(&result).Error
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find item by call ID")
	}
	return result.EtoD(), nil
}

// GetItemByCallIDAndType implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) GetItemByCallIDAndType(ctx context.Context, conversationID uint, callID string, itemType conversation.ItemType) (*conversation.Item, error) {
	q := repo.db.GetQuery(ctx)
	// Use raw SQL since gormgen may not have the call_id field
	var result dbschema.ConversationItem
	err := q.ConversationItem.WithContext(ctx).UnderlyingDB().
		Where("conversation_id = ? AND call_id = ? AND type = ?", conversationID, callID, string(itemType)).
		First(&result).Error
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find item by call ID and type")
	}
	return result.EtoD(), nil
}

// UpdateItem implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) UpdateItem(ctx context.Context, conversationID uint, item *conversation.Item) error {
	q := repo.db.GetQuery(ctx)
	entity := dbschema.NewSchemaConversationItem(item)
	
	_, err := q.ConversationItem.WithContext(ctx).
		Where(q.ConversationItem.ID.Eq(item.ID)).
		Where(q.ConversationItem.ConversationID.Eq(conversationID)).
		Updates(entity)
	if err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to update item")
	}
	return nil
}

// DeleteItem implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) DeleteItem(ctx context.Context, conversationID uint, itemID uint) error {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyItemFilter(q, sql, conversation.ItemFilter{
		ID:             &itemID,
		ConversationID: &conversationID,
	})
	_, err := sql.Delete()
	if err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to delete item")
	}
	return nil
}

// CountItems implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) CountItems(ctx context.Context, conversationID uint, branchName string) (int, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyItemFilter(q, sql, conversation.ItemFilter{
		ConversationID: &conversationID,
	})

	// For now, we count all items since branch filtering isn't fully implemented in gormgen
	count, err := sql.Count()

	if err != nil {
		return 0, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to count items")
	}

	return int(count), nil
}

// Branch operations
// CreateBranch implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) CreateBranch(ctx context.Context, conversationID uint, branchName string, metadata *conversation.BranchMetadata) error {
	// Verify conversation exists
	_, err := repo.FindByID(ctx, conversationID)
	if err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "conversation not found")
	}

	// TODO: Implement branch storage in database
	// For now, return not implemented error
	return platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "branch operations not yet implemented in database layer", nil, "b4c5d6e7-f8a9-4b0c-1d2e-3f4a5b6c7d8e")
}

// GetBranch implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) GetBranch(ctx context.Context, conversationID uint, branchName string) (*conversation.BranchMetadata, error) {
	return nil, platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "branch operations not yet implemented in database layer", nil, "c5d6e7f8-a9b0-4c1d-2e3f-4a5b6c7d8e9f")
}

// ListBranches implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) ListBranches(ctx context.Context, conversationID uint) ([]*conversation.BranchMetadata, error) {
	return nil, platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "branch operations not yet implemented in database layer", nil, "d6e7f8a9-b0c1-4d2e-3f4a-5b6c7d8e9f0a")
}

// DeleteBranch implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) DeleteBranch(ctx context.Context, conversationID uint, branchName string) error {
	return platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "branch operations not yet implemented in database layer", nil, "e7f8a9b0-c1d2-4e3f-4a5b-6c7d8e9f0a1b")
}

// SetActiveBranch implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) SetActiveBranch(ctx context.Context, conversationID uint, branchName string) error {
	return platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "branch operations not yet implemented in database layer", nil, "f8a9b0c1-d2e3-4f4a-5b6c-7d8e9f0a1b2c")
}

// Branch item operations
// AddItemToBranch implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) AddItemToBranch(ctx context.Context, conversationID uint, branchName string, item *conversation.Item) error {
	// For now, branch operations are not implemented
	// Default to MAIN branch behavior
	if branchName == "MAIN" || branchName == "" {
		return repo.AddItem(ctx, conversationID, item)
	}
	return platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "branch operations not yet implemented in database layer", nil, "a9b0c1d2-e3f4-4a5b-6c7d-8e9f0a1b2c3d")
}

// GetBranchItems implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) GetBranchItems(ctx context.Context, conversationID uint, branchName string, pagination *query.Pagination) ([]*conversation.Item, error) {
	// For now, return items for MAIN branch with pagination support
	if branchName == "MAIN" || branchName == "" {
		q := repo.db.GetQuery(ctx)
		sql := q.ConversationItem.WithContext(ctx)
		sql = repo.applyItemFilter(q, sql, conversation.ItemFilter{
			ConversationID: &conversationID,
		})
		sql = repo.applyItemPagination(q, sql, pagination)

		rows, err := sql.Find()
		if err != nil {
			return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to get branch items")
		}

		return functional.Map(rows, func(item *dbschema.ConversationItem) *conversation.Item {
			return item.EtoD()
		}), nil
	}
	return nil, platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "branch operations not yet implemented in database layer", nil, "b0c1d2e3-f4a5-4b6c-7d8e-9f0a1b2c3d4e")
}

// applyItemPagination applies pagination to item queries
func (repo *ConversationGormRepository) applyItemPagination(q *gormgen.Query, sql gormgen.IConversationItemDo, p *query.Pagination) gormgen.IConversationItemDo {
	if p != nil {
		// Apply cursor-based pagination
		if p.After != nil {
			if p.Order == "desc" {
				sql = sql.Where(q.ConversationItem.ID.Lt(*p.After))
			} else {
				sql = sql.Where(q.ConversationItem.ID.Gt(*p.After))
			}
		}

		// Apply ordering (default to ascending by ID)
		if p.Order == "desc" {
			sql = sql.Order(q.ConversationItem.ID.Desc())
		} else {
			sql = sql.Order(q.ConversationItem.ID.Asc())
		}

		// Apply limit
		if p.Limit != nil && *p.Limit > 0 {
			sql = sql.Limit(*p.Limit)
		}
	} else {
		// Default ordering when no pagination specified
		sql = sql.Order(q.ConversationItem.ID.Asc())
	}
	return sql
}

// BulkAddItemsToBranch implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) BulkAddItemsToBranch(ctx context.Context, conversationID uint, branchName string, items []*conversation.Item) error {
	// For now, branch operations are not implemented
	// Default to MAIN branch behavior
	if branchName == "MAIN" || branchName == "" {
		return repo.BulkAddItems(ctx, conversationID, items)
	}
	return platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "branch operations not yet implemented in database layer", nil, "c1d2e3f4-a5b6-4c7d-8e9f-0a1b2c3d4e5f")
}

// ForkBranch implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) ForkBranch(ctx context.Context, conversationID uint, sourceBranch, newBranch string, fromItemID string, description *string) error {
	return platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "branch operations not yet implemented in database layer", nil, "d2e3f4a5-b6c7-4d8e-9f0a-1b2c3d4e5f6a")
}

// Item rating operations
// RateItem implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) RateItem(ctx context.Context, conversationID uint, itemID string, rating conversation.ItemRating, comment *string) error {
	// TODO: Implement rating storage in database
	// For now, return not implemented error
	return platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "rating operations not yet implemented in database layer", nil, "e3f4a5b6-c7d8-4e9f-0a1b-2c3d4e5f6a7b")
}

// GetItemRating implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) GetItemRating(ctx context.Context, conversationID uint, itemID string) (*conversation.ItemRating, error) {
	return nil, platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "rating operations not yet implemented in database layer", nil, "f4a5b6c7-d8e9-4f0a-1b2c-3d4e5f6a7b8c")
}

// RemoveItemRating implements conversation.ConversationRepository.
func (repo *ConversationGormRepository) RemoveItemRating(ctx context.Context, conversationID uint, itemID string) error {
	return platformerrors.NewError(ctx, platformerrors.LayerRepository, platformerrors.ErrorTypeNotImplemented, "rating operations not yet implemented in database layer", nil, "a5b6c7d8-e9f0-4a1b-2c3d-4e5f6a7b8c9d")
}

// applyFilter applies filter conditions to the query
func (repo *ConversationGormRepository) applyFilter(q *gormgen.Query, sql gormgen.IConversationDo, filter conversation.ConversationFilter) gormgen.IConversationDo {
	if filter.ID != nil {
		sql = sql.Where(q.Conversation.ID.Eq(*filter.ID))
	}
	if filter.PublicID != nil {
		sql = sql.Where(q.Conversation.PublicID.Eq(*filter.PublicID))
	}
	if filter.UserID != nil {
		sql = sql.Where(q.Conversation.UserID.Eq(*filter.UserID))
	}
	if filter.Referrer != nil && *filter.Referrer != "" {
		sql = sql.Where(q.Conversation.Referrer.Eq(*filter.Referrer))
	}
	return sql
}

// applyItemFilter applies filter conditions to the conversation item query
func (repo *ConversationGormRepository) applyItemFilter(q *gormgen.Query, sql gormgen.IConversationItemDo, filter conversation.ItemFilter) gormgen.IConversationItemDo {
	if filter.ID != nil {
		sql = sql.Where(q.ConversationItem.ID.Eq(*filter.ID))
	}
	if filter.PublicID != nil {
		sql = sql.Where(q.ConversationItem.PublicID.Eq(*filter.PublicID))
	}
	// Note: CallID filtering is done via raw SQL in GetItemByCallID since gormgen may not have the field
	if filter.ConversationID != nil {
		sql = sql.Where(q.ConversationItem.ConversationID.Eq(*filter.ConversationID))
	}
	if filter.Role != nil {
		roleStr := string(*filter.Role)
		sql = sql.Where(q.ConversationItem.Role.Eq(roleStr))
	}
	if filter.ResponseID != nil {
		sql = sql.Where(q.ConversationItem.ResponseID.Eq(*filter.ResponseID))
	}
	return sql
}

// applyPagination applies pagination to the query
func (repo *ConversationGormRepository) applyPagination(q *gormgen.Query, sql gormgen.IConversationDo, p *query.Pagination) gormgen.IConversationDo {
	if p != nil {
		if p.Limit != nil && *p.Limit > 0 {
			sql = sql.Limit(*p.Limit)
		}
		if p.After != nil {
			if p.Order == "desc" {
				sql = sql.Where(q.Conversation.ID.Lt(*p.After))
			} else {
				sql = sql.Where(q.Conversation.ID.Gt(*p.After))
			}
		}
		if p.Order == "desc" {
			sql = sql.Order(q.Conversation.ID.Desc())
		} else {
			sql = sql.Order(q.Conversation.ID.Asc())
		}
	}
	return sql
}
