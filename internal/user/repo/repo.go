package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"avito_2023/internal/database"
	sModel "avito_2023/internal/segment/model"
	"avito_2023/internal/user/model"
)

//go:generate moq --out mocks/repo_mock.go --pkg=mocks . Repo

type Repo interface {
	// GetUserSegments - get user segments
	GetUserSegments(ctx context.Context, userID uint) ([]*string, error)

	// GetUserHistory - get user history
	GetUserHistory(ctx context.Context, userID uint, month uint, year uint) ([]*model.UserHistory, error)

	// UpdateUserSegments - update user segment
	UpdateUserSegments(ctx context.Context, userID uint, slugsToAdd, slugsToDel []string, deleteAt *time.Time) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) Repo {
	return &repo{
		db: db,
	}
}

func (r *repo) GetUserSegments(ctx context.Context, userID uint) ([]*string, error) {
	db := database.FromContext(ctx, r.db)

	var segments []*string
	if err := db.WithContext(ctx).
		Model(&model.UserSegmentDB{}).
		Select("slug").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL OR deleted_at > NOW()").
		Joins("LEFT JOIN segments ON users_segments.segment_id = segments.id").
		Scan(&segments).Error; err != nil {
		return nil, err
	}
	if len(segments) == 0 {
		return nil, database.ErrNotFound
	}

	return segments, nil
}

func (r *repo) GetUserHistory(ctx context.Context, userID uint, month uint, year uint) ([]*model.UserHistory, error) {
	db := database.FromContext(ctx, r.db)

	var history []*model.UserHistory
	if err := db.WithContext(ctx).
		Model(&model.UserSegmentDB{}).
		Select("slug", "created_at", "deleted_at").
		Where("user_id = ? AND EXTRACT(MONTH FROM created_at) = ? AND EXTRACT(YEAR FROM created_at) = ?", userID, month, year).
		Joins("LEFT JOIN segments ON users_segments.segment_id = segments.id").Scan(&history).Error; err != nil {
		return nil, err
	}
	if len(history) == 0 {
		return nil, database.ErrNotFound
	}

	return history, nil
}

func (r *repo) UpdateUserSegments(ctx context.Context, userID uint, slugsToAdd, slugsToDel []string, deleteAt *time.Time) error {
	db := database.FromContext(ctx, r.db)

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user *model.UserDB
		if err := tx.Where(model.UserDB{ID: userID}).
			FirstOrCreate(&user).Error; err != nil {
			return err
		}

		if len(slugsToAdd) != 0 {
			var idsToAdd []uint
			if err := tx.Model(&sModel.SegmentDB{}).
				Select("id").
				Where("slug IN ?", slugsToAdd).
				Find(&idsToAdd).Error; err != nil {
				return err
			}
			if len(idsToAdd) == 0 {
				return database.ErrUpdateUserSegments_InvalidSegments
			}

			userSegments := make([]*model.UserSegmentDB, 0, len(idsToAdd))
			for _, id := range idsToAdd {
				row := &model.UserSegmentDB{UserID: userID, SegmentID: id}
				if deleteAt != nil {
					row.DeletedAt = deleteAt
				}

				userSegments = append(userSegments, row)
			}
			if err := tx.Model(&model.UserSegmentDB{}).Create(&userSegments).Error; err != nil {
				return err
			}
		}

		if len(slugsToDel) != 0 {
			var idsToDel []uint
			if err := tx.Model(&sModel.SegmentDB{}).
				Select("id").
				Where("slug IN ?", slugsToDel).
				Find(&idsToDel).Error; err != nil {
				return err
			}
			if len(idsToDel) == 0 {
				return nil
			}
			if err := tx.Model(&model.UserSegmentDB{}).Where("user_id = ? AND segment_id IN ?", userID, idsToDel).
				Update("deleted_at", "NOW()").Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
