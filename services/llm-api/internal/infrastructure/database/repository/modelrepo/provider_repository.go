package modelrepo

import (
	"context"

	domainmodel "jan-server/services/llm-api/internal/domain/model"
	"jan-server/services/llm-api/internal/domain/query"
	"jan-server/services/llm-api/internal/infrastructure/database/dbschema"
	"jan-server/services/llm-api/internal/infrastructure/database/gormgen"
	"jan-server/services/llm-api/internal/infrastructure/database/transaction"
	"jan-server/services/llm-api/internal/utils/functional"
)

type ProviderGormRepository struct {
	db *transaction.Database
}

var _ domainmodel.ProviderRepository = (*ProviderGormRepository)(nil)

func NewProviderGormRepository(db *transaction.Database) domainmodel.ProviderRepository {
	return &ProviderGormRepository{db: db}
}

func (repo *ProviderGormRepository) applyFilter(query *gormgen.Query, sql gormgen.IProviderDo, filter domainmodel.ProviderFilter) gormgen.IProviderDo {
	if filter.IDs != nil && len(*filter.IDs) > 0 {
		sql = sql.Where(query.Provider.ID.In((*filter.IDs)...))
	}
	if filter.PublicID != nil {
		sql = sql.Where(query.Provider.PublicID.Eq(*filter.PublicID))
	}
	if filter.Kind != nil {
		sql = sql.Where(query.Provider.Kind.Eq(string(*filter.Kind)))
	}
	if filter.Active != nil {
		sql = sql.Where(query.Provider.Active.Is(*filter.Active))
	}
	if filter.IsModerated != nil {
		sql = sql.Where(query.Provider.IsModerated.Is(*filter.IsModerated))
	}
	if filter.LastSyncedAfter != nil {
		sql = sql.Where(query.Provider.LastSyncedAt.Gte(*filter.LastSyncedAfter))
	}
	if filter.LastSyncedBefore != nil {
		sql = sql.Where(query.Provider.LastSyncedAt.Lte(*filter.LastSyncedBefore))
	}
	return sql
}

func (repo *ProviderGormRepository) Create(ctx context.Context, provider *domainmodel.Provider) error {
	model := dbschema.NewSchemaProvider(provider)
	query := repo.db.GetQuery(ctx)
	if err := query.Provider.WithContext(ctx).Create(model); err != nil {
		return err
	}
	provider.ID = model.ID
	provider.CreatedAt = model.CreatedAt
	provider.UpdatedAt = model.UpdatedAt
	return nil
}

func (repo *ProviderGormRepository) Update(ctx context.Context, provider *domainmodel.Provider) error {
	model := dbschema.NewSchemaProvider(provider)
	query := repo.db.GetQuery(ctx)
	_, err := query.Provider.WithContext(ctx).Where(query.Provider.ID.Eq(model.ID)).Updates(model)
	return err
}

func (repo *ProviderGormRepository) DeleteByID(ctx context.Context, id uint) error {
	query := repo.db.GetQuery(ctx)
	_, err := query.Provider.WithContext(ctx).
		Where(query.Provider.ID.Eq(id)).
		Delete(&dbschema.Provider{})
	return err
}

func (repo *ProviderGormRepository) FindByID(ctx context.Context, id uint) (*domainmodel.Provider, error) {
	ids := []uint{id}
	filter := domainmodel.ProviderFilter{
		IDs: &ids,
	}
	results, err := repo.FindByFilter(ctx, filter, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (repo *ProviderGormRepository) FindByPublicID(ctx context.Context, publicID string) (*domainmodel.Provider, error) {
	filter := domainmodel.ProviderFilter{
		PublicID: &publicID,
	}
	results, err := repo.FindByFilter(ctx, filter, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (repo *ProviderGormRepository) FindByFilter(ctx context.Context, filter domainmodel.ProviderFilter, p *query.Pagination) ([]*domainmodel.Provider, error) {
	query := repo.db.GetQuery(ctx)
	sql := query.Provider.WithContext(ctx)
	sql = repo.applyFilter(query, sql, filter)
	if p != nil {
		if p.Limit != nil && *p.Limit > 0 {
			sql = sql.Limit(*p.Limit)
		}
		if p.Offset != nil && *p.Offset >= 0 {
			sql = sql.Offset(*p.Offset)
		}
		if p.Order == "desc" {
			sql = sql.Order(query.Provider.CreatedAt.Desc())
		} else {
			sql = sql.Order(query.Provider.CreatedAt.Asc())
		}
	}
	rows, err := sql.Find()
	if err != nil {
		return nil, err
	}
	providers := functional.Map(rows, func(item *dbschema.Provider) *domainmodel.Provider {
		return item.EtoD()
	})
	return providers, nil
}

func (repo *ProviderGormRepository) Count(ctx context.Context, filter domainmodel.ProviderFilter) (int64, error) {
	query := repo.db.GetQuery(ctx)
	sql := query.Provider.WithContext(ctx)
	sql = repo.applyFilter(query, sql, filter)
	return sql.Count()
}

func (repo *ProviderGormRepository) FindByIDs(ctx context.Context, ids []uint) ([]*domainmodel.Provider, error) {
	if len(ids) == 0 {
		return []*domainmodel.Provider{}, nil
	}

	filter := domainmodel.ProviderFilter{
		IDs: &ids,
	}
	return repo.FindByFilter(ctx, filter, nil)
}
