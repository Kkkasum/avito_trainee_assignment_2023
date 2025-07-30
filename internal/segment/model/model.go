package model

type SegmentDB struct {
	ID   uint   `gorm:"id"`
	Slug string `gorm:"slug"`
}

func (SegmentDB) TableName() string {
	return "segments"
}
