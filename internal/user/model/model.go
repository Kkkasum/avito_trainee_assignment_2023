package model

import (
	"time"
)

type UserDB struct {
	ID        uint      `gorm:"id"`
	CreatedAt time.Time `gorm:"created_at"`
}

func (UserDB) TableName() string {
	return "users"
}

type UserSegmentDB struct {
	ID        uint       `gorm:"id"`
	UserID    uint       `gorm:"user_id"`
	SegmentID uint       `gorm:"segment_id"`
	CreatedAt time.Time  `gorm:"created_at"`
	DeletedAt *time.Time `gorm:"deleted_at"`
}

func (UserSegmentDB) TableName() string {
	return "users_segments"
}

type UserHistory struct {
	Slug      string     `gorm:"slug" json:"slug"`
	CreatedAt time.Time  `gorm:"created_at" json:"created_at"`
	DeletedAt *time.Time `gorm:"deleted_at" json:"deleted_at"`
}
