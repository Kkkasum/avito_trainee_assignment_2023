package database

import (
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNotFound                           = errors.New("record not found")
	ErrUpdateUserSegments_InvalidSegments = errors.New("invalid segments")
)

func IsRecordNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows)
}

func IsUpdateUserSegmentsInvalidSegmentsErr(err error) bool {
	return errors.Is(err, ErrUpdateUserSegments_InvalidSegments)
}
