package db

import (
	"github.com/opoccomaxao/ffmpegbox/internal/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Migrate(
	db *gorm.DB,
) error {
	err := db.AutoMigrate(
		&models.Task{},
	)
	if err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}

	return nil
}
