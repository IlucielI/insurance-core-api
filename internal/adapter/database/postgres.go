package database

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresConfig struct {
	DatabaseURL     string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
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

	maxOpen := config.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 10
	}
	maxIdle := config.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 5
	}
	maxLifetime := config.ConnMaxLifetime
	if maxLifetime <= 0 {
		maxLifetime = 30 * time.Minute
	}
	maxIdleTime := config.ConnMaxIdleTime
	if maxIdleTime <= 0 {
		maxIdleTime = 10 * time.Minute
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(maxLifetime)
	sqlDB.SetConnMaxIdleTime(maxIdleTime)

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
