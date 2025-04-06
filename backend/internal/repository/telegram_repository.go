package repository

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"

	"goout/config"
)

type TelegramRepository struct {
	dialector gorm.Dialector
	db        *gorm.DB
}

func NewTelegramRepository(config *config.Config) (*TelegramRepository, error) {
	dialector := sqlite.Open(config.DBPath)

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	driver, err := sqlite3.WithInstance(sqlDB, &sqlite3.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+config.MigrationsPath,
		"sqlite3", driver)
	if err != nil {
		return nil, fmt.Errorf("could not start migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &TelegramRepository{
		dialector: dialector,
		db:        db,
	}, nil
}

func (r *TelegramRepository) GetDialector() gorm.Dialector {
	return r.dialector
}

func (r *TelegramRepository) Stop() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
