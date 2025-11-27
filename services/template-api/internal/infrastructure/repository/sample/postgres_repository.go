package sample

import (
	"context"
	"errors"

	"gorm.io/gorm"

	domain "jan-server/services/template-api/internal/domain/sample"
	"jan-server/services/template-api/internal/infrastructure/database/entities"
	"jan-server/services/template-api/internal/utils/platformerrors"
)

// PostgresRepository persists samples via PostgreSQL using GORM.
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a repository backed by the provided DB.
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// FetchLatest returns the most recently created sample row.
func (r *PostgresRepository) FetchLatest(ctx context.Context) (domain.Sample, error) {
	var record entities.Sample
	err := r.db.WithContext(ctx).Order("created_at DESC").First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Sample{}, platformerrors.NewError(
				ctx,
				platformerrors.LayerRepository,
				platformerrors.ErrorTypeNotFound,
				"no sample records found",
				err,
				"3e4f5a6b-7c8d-4e9f-0a1b-2c3d4e5f6a7b",
			)
		}
		return domain.Sample{}, platformerrors.NewError(
			ctx,
			platformerrors.LayerRepository,
			platformerrors.ErrorTypeDatabaseError,
			"failed to fetch latest sample",
			err,
			"4f5a6b7c-8d9e-4f0a-1b2c-3d4e5f6a7b8c",
		)
	}

	return domain.Sample{
		ID:      record.ID,
		Message: record.Message,
	}, nil
}
