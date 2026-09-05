package database

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migration struct {
	Name string
	SQL  string
}

func RunMigrations(db *gorm.DB) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return errors.New("no migrations found")
	}

	for _, migration := range migrations {
		if err := db.Exec(migration.SQL).Error; err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Name, err)
		}
	}

	return nil
}

func loadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := "migrations/" + entry.Name()
		content, err := fs.ReadFile(migrationsFS, path)
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, Migration{
			Name: entry.Name(),
			SQL:  string(content),
		})
	}

	sort.Slice(migrations, func(left int, right int) bool {
		return migrations[left].Name < migrations[right].Name
	})

	return migrations, nil
}
