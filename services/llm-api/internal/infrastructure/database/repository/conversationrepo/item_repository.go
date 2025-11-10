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

type ItemGormRepository struct {
	db *transaction.Database
}

var _ conversation.ItemRepository = (*ItemGormRepository)(nil)

func NewItemGormRepository(db *transaction.Database) conversation.ItemRepository {
	return &ItemGormRepository{db}
}

// Create implements conversation.ItemRepository.
func (repo *ItemGormRepository) Create(ctx context.Context, item *conversation.Item) error {
	model := dbschema.NewSchemaConversationItem(item)
	if err := repo.db.GetQuery(ctx).ConversationItem.WithContext(ctx).Create(model); err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to create item")
	}
	// Update the domain object with generated ID and timestamps
	item.ID = model.ID
	item.CreatedAt = model.CreatedAt
	return nil
}

// FindByID implements conversation.ItemRepository.
func (repo *ItemGormRepository) FindByID(ctx context.Context, id uint) (*conversation.Item, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyFilter(q, sql, conversation.ItemFilter{ID: &id})
	result, err := sql.First()
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find item by ID")
	}
	return result.EtoD(), nil
}

// FindByPublicID implements conversation.ItemRepository.
func (repo *ItemGormRepository) FindByPublicID(ctx context.Context, publicID string) (*conversation.Item, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyFilter(q, sql, conversation.ItemFilter{PublicID: &publicID})
	result, err := sql.First()
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find item by public ID")
	}
	return result.EtoD(), nil
}

// FindByConversationID implements conversation.ItemRepository.
func (repo *ItemGormRepository) FindByConversationID(ctx context.Context, conversationID uint) ([]*conversation.Item, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyFilter(q, sql, conversation.ItemFilter{ConversationID: &conversationID})
	rows, err := sql.Order(q.ConversationItem.CreatedAt.Asc()).Find()
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find items by conversation ID")
	}

	result := functional.Map(rows, func(item *dbschema.ConversationItem) *conversation.Item {
		return item.EtoD()
	})
	return result, nil
}

// Search implements conversation.ItemRepository.
func (repo *ItemGormRepository) Search(ctx context.Context, conversationID uint, searchQuery string) ([]*conversation.Item, error) {
	// For now, this is a simple implementation
	// In production, you'd want to use full-text search or a search engine
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyFilter(q, sql, conversation.ItemFilter{ConversationID: &conversationID})

	// Basic search - in production, enhance with proper full-text search
	rows, err := sql.Order(q.ConversationItem.CreatedAt.Asc()).Find()

	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to search items")
	}

	result := functional.Map(rows, func(item *dbschema.ConversationItem) *conversation.Item {
		return item.EtoD()
	})

	// TODO: Implement proper full-text search filtering based on searchQuery
	return result, nil
}

// Delete implements conversation.ItemRepository.
func (repo *ItemGormRepository) Delete(ctx context.Context, id uint) error {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyFilter(q, sql, conversation.ItemFilter{ID: &id})
	_, err := sql.Delete()
	if err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to delete item")
	}
	return nil
}

// BulkCreate implements conversation.ItemRepository.
func (repo *ItemGormRepository) BulkCreate(ctx context.Context, items []*conversation.Item) error {
	if len(items) == 0 {
		return nil
	}

	// Convert to schema models
	models := functional.Map(items, func(item *conversation.Item) *dbschema.ConversationItem {
		return dbschema.NewSchemaConversationItem(item)
	})

	// Bulk insert
	q := repo.db.GetQuery(ctx)
	if err := q.ConversationItem.WithContext(ctx).CreateInBatches(models, 100); err != nil {
		return platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to bulk create items")
	}

	// Update domain objects with generated IDs
	for i, model := range models {
		items[i].ID = model.ID
		items[i].CreatedAt = model.CreatedAt
	}

	return nil
}

// CountByConversation implements conversation.ItemRepository.
func (repo *ItemGormRepository) CountByConversation(ctx context.Context, conversationID uint) (int64, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyFilter(q, sql, conversation.ItemFilter{ConversationID: &conversationID})
	count, err := sql.Count()
	if err != nil {
		return 0, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to count items by conversation")
	}
	return count, nil
}

// ExistsByIDAndConversation implements conversation.ItemRepository.
func (repo *ItemGormRepository) ExistsByIDAndConversation(ctx context.Context, itemID uint, conversationID uint) (bool, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyFilter(q, sql, conversation.ItemFilter{
		ID:             &itemID,
		ConversationID: &conversationID,
	})
	count, err := sql.Count()
	if err != nil {
		return false, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to check item existence")
	}
	return count > 0, nil
}

// FindByFilter implements conversation.ItemRepository.
func (repo *ItemGormRepository) FindByFilter(ctx context.Context, filter conversation.ItemFilter, pagination *query.Pagination) ([]*conversation.Item, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyFilter(q, sql, filter)
	sql = repo.applyPagination(q, sql, pagination)

	rows, err := sql.Find()
	if err != nil {
		return nil, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to find items by filter")
	}

	result := functional.Map(rows, func(item *dbschema.ConversationItem) *conversation.Item {
		return item.EtoD()
	})
	return result, nil
}

// Count implements conversation.ItemRepository.
func (repo *ItemGormRepository) Count(ctx context.Context, filter conversation.ItemFilter) (int64, error) {
	q := repo.db.GetQuery(ctx)
	sql := q.ConversationItem.WithContext(ctx)
	sql = repo.applyFilter(q, sql, filter)
	count, err := sql.Count()
	if err != nil {
		return 0, platformerrors.AsError(ctx, platformerrors.LayerRepository, err, "failed to count items")
	}
	return count, nil
}

// applyFilter applies filter conditions to the query
func (repo *ItemGormRepository) applyFilter(q *gormgen.Query, sql gormgen.IConversationItemDo, filter conversation.ItemFilter) gormgen.IConversationItemDo {
	if filter.PublicID != nil {
		sql = sql.Where(q.ConversationItem.PublicID.Eq(*filter.PublicID))
	}
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
func (repo *ItemGormRepository) applyPagination(q *gormgen.Query, sql gormgen.IConversationItemDo, p *query.Pagination) gormgen.IConversationItemDo {
	if p != nil {
		if p.Limit != nil && *p.Limit > 0 {
			sql = sql.Limit(*p.Limit)
		}
		if p.After != nil {
			if p.Order == "desc" {
				sql = sql.Where(q.ConversationItem.ID.Lt(*p.After))
			} else {
				sql = sql.Where(q.ConversationItem.ID.Gt(*p.After))
			}
		}
		if p.Order == "desc" {
			sql = sql.Order(q.ConversationItem.ID.Desc())
		} else {
			sql = sql.Order(q.ConversationItem.ID.Asc())
		}
	}
	return sql
}
