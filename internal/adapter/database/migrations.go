package database

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	migrationfiles "github.com/bayuanugerah/insurance-core-api/migrations"
	"gorm.io/gorm"
)

type Migration struct {
	Name string
	SQL  string
}

func RunMigrations(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(734821)).Error; err != nil {
			return fmt.Errorf("failed to acquire migration lock: %w", err)
		}
		return runMigrations(tx)
	})
}

func runMigrations(db *gorm.DB) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return errors.New("no migrations found")
	}

	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`).Error; err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	for _, migration := range migrations {
		var applied bool
		if err := db.Raw("SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name = ?)", migration.Name).Scan(&applied).Error; err != nil {
			return fmt.Errorf("failed to check migration %s: %w", migration.Name, err)
		}
		if applied {
			continue
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(migration.SQL).Error; err != nil {
				return err
			}
			return tx.Exec("INSERT INTO schema_migrations (name) VALUES (?)", migration.Name).Error
		}); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Name, err)
		}
	}

	return nil
}

func loadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationfiles.FS, ".")
	if err != nil {
		return nil, err
	}

	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := fs.ReadFile(migrationfiles.FS, entry.Name())
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, Migration{Name: entry.Name(), SQL: string(content)})
	}

	sort.Slice(migrations, func(left int, right int) bool {
		return migrations[left].Name < migrations[right].Name
	})

	return migrations, nil
}
