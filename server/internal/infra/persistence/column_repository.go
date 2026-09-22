package persistence

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/shawns-yao/shawn-blog/server/internal/domain/content"
	"github.com/shawns-yao/shawn-blog/server/internal/infra/persistence/model"
)

type MomentColumnRepository struct {
	db   *gorm.DB
	repo *GormRepository[model.MomentColumn]
}

func NewMomentColumnRepository(db *gorm.DB) *MomentColumnRepository {
	return &MomentColumnRepository{
		db:   db,
		repo: NewGormRepository[model.MomentColumn](db),
	}
}

func (r *MomentColumnRepository) List(ctx context.Context) ([]*content.MomentColumn, error) {
	records, err := r.repo.List(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	})
	if err != nil {
		return nil, err
	}
	result := make([]*content.MomentColumn, len(records))
	for i, rec := range records {
		item := mapColumnToDomain(rec)
		result[i] = item
	}
	return result, nil
}

func (r *MomentColumnRepository) GetByID(ctx context.Context, id int64) (*content.MomentColumn, error) {
	rec, err := r.repo.FirstByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, content.ErrColumnNotFound
		}
		return nil, err
	}
	return mapColumnToDomain(*rec), nil
}

func (r *MomentColumnRepository) Create(ctx context.Context, column *content.MomentColumn) error {
	rec := model.MomentColumn{
		ParentID: column.ParentID,
		Name:     column.Name,
		ShortURL: optionalString(column.ShortURL),
	}
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockColumnHierarchy(tx); err != nil {
			return err
		}
		if err := validateColumnParent(tx, column); err != nil {
			return err
		}
		return tx.Create(&rec).Error
	}); err != nil {
		return err
	}
	column.ID = rec.ID
	column.CreatedAt = rec.CreatedAt
	column.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *MomentColumnRepository) Update(ctx context.Context, column *content.MomentColumn) error {
	updates := map[string]any{
		"parent_id":  column.ParentID,
		"name":       column.Name,
		"short_url":  optionalString(column.ShortURL),
		"updated_at": time.Now(),
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockColumnHierarchy(tx); err != nil {
			return err
		}
		if err := validateColumnParent(tx, column); err != nil {
			return err
		}
		result := tx.Model(&model.MomentColumn{}).Where("id = ?", column.ID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return content.ErrColumnNotFound
		}
		return nil
	})
}

func (r *MomentColumnRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockColumnHierarchy(tx); err != nil {
			return err
		}
		var children, articles int64
		if err := tx.Model(&model.MomentColumn{}).Where("parent_id = ?", id).Count(&children).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Moment{}).Where("column_id = ?", id).Count(&articles).Error; err != nil {
			return err
		}
		if children > 0 || articles > 0 {
			return content.ErrColumnInUse
		}
		result := tx.Where("id = ?", id).Delete(&model.MomentColumn{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return content.ErrColumnNotFound
		}
		return nil
	})
}

// Serialize category edits so concurrent moves cannot create a third level or a cycle.
func lockColumnHierarchy(tx *gorm.DB) error {
	return tx.Exec("LOCK TABLE moment_column IN SHARE ROW EXCLUSIVE MODE").Error
}

func validateColumnParent(tx *gorm.DB, column *content.MomentColumn) error {
	if column.ParentID == nil {
		return nil
	}
	if *column.ParentID <= 0 || *column.ParentID == column.ID {
		return content.ErrColumnHierarchy
	}
	var parent model.MomentColumn
	if err := tx.First(&parent, *column.ParentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return content.ErrColumnHierarchy
		}
		return err
	}
	if parent.ParentID != nil {
		return content.ErrColumnHierarchy
	}
	if column.ID > 0 {
		var children int64
		if err := tx.Model(&model.MomentColumn{}).Where("parent_id = ?", column.ID).Count(&children).Error; err != nil {
			return err
		}
		if children > 0 {
			return content.ErrColumnHierarchy
		}
	}
	return nil
}

func mapColumnToDomain(rec model.MomentColumn) *content.MomentColumn {
	return &content.MomentColumn{
		ParentID:  rec.ParentID,
		ID:        rec.ID,
		Name:      rec.Name,
		ShortURL:  stringToPtr(rec.ShortURL),
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
		DeletedAt: deletedAtToPtr(rec.DeletedAt),
	}
}
