package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestNewPostgresRequiresDatabaseURL(t *testing.T) {
	_, err := NewPostgres(PostgresConfig{})
	if err == nil || err.Error() != "DatabaseURL is required" {
		t.Fatalf("NewPostgres() error = %v, want DatabaseURL required", err)
	}
}

func TestPostgresDBAndClose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	postgresDB := &Postgres{db: gormDB, sqlDB: db}

	if postgresDB.DB() != gormDB {
		t.Fatal("DB() did not return wrapped gorm DB")
	}
	mock.ExpectClose()
	if err := postgresDB.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestCloseGormDB(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	closeGormDB(gormDB)
}
