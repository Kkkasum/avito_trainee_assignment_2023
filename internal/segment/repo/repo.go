package repo

import (
	"context"
	
	"gorm.io/gorm"

	"avito_2023/internal/database"
	"avito_2023/internal/segment/model"
	uModel "avito_2023/internal/user/model"
)

//go:generate moq --out mocks/repo_mock.go --pkg=mocks . Repo

type Repo interface {
	// AddSegment - add new segment
	AddSegment(ctx context.Context, slug string, percentage uint) error

	// DeleteSegment - delete segment
	DeleteSegment(ctx context.Context, slug string) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) Repo {
	return &repo{
		db: db,
	}
}

func (r *repo) AddSegment(ctx context.Context, slug string, percentage uint) error {
	db := database.FromContext(ctx, r.db)

	if err := db.Transaction(func(tx *gorm.DB) error {
		newSegment := &model.SegmentDB{Slug: slug}
		if err := tx.Create(newSegment).Error; err != nil {
			return err
		}

		if percentage == 0 {
			return nil
		}

		// get random users ids
		var usersCount int64
		if err := tx.Model(&uModel.UserDB{}).
			Count(&usersCount).Error; err != nil {
			return err
		}
		limit := int(usersCount * int64(percentage) / 100)

		var usersIDs []uint
		if err := tx.Model(&uModel.UserDB{}).
			Select("id").
			Order("RANDOM()").
			Limit(limit).
			Find(&usersIDs).Error; err != nil {
			return err
		}

		newUsersSegments := make([]*uModel.UserSegmentDB, len(usersIDs))
		for i, userID := range usersIDs {
			newUsersSegments[i] = &uModel.UserSegmentDB{
				UserID:    userID,
				SegmentID: newSegment.ID,
			}
		}
		if err := tx.Model(&uModel.UserSegmentDB{}).Create(&newUsersSegments).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *repo) DeleteSegment(ctx context.Context, slug string) error {
	db := database.FromContext(ctx, r.db)

	res := db.WithContext(ctx).
		Delete(&model.SegmentDB{}, "slug = ?", slug)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}
