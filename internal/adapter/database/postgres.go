package database

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConfig struct {
	DatabaseURL string
}

type Postgres struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

func NewPostgres(config PostgresConfig) (*Postgres, error) {
	if config.DatabaseURL == "" {
		return nil, errors.New("DatabaseURL is required")
	}

	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		closeGormDB(db)
		return nil, err
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		closeGormDB(db)
		return nil, err
	}

	return &Postgres{
		db:    db,
		sqlDB: sqlDB,
	}, nil
}

func (postgres *Postgres) DB() *gorm.DB {
	return postgres.db
}

func (postgres *Postgres) Close() error {
	return postgres.sqlDB.Close()
}

func closeGormDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}

	_ = sqlDB.Close()
}
