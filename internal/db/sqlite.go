package db

import (
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	DSN string
}

func New(config *Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(config.DSN), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get underlying database")
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	return db, nil
}
