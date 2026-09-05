package database

import "testing"

func TestLoadMigrations(t *testing.T) {
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}
	if len(migrations) < 6 {
		t.Fatalf("loadMigrations() length = %d, want at least 6", len(migrations))
	}
	for index := 1; index < len(migrations); index++ {
		if migrations[index-1].Name > migrations[index].Name {
			t.Fatalf("migrations are not sorted: %s before %s", migrations[index-1].Name, migrations[index].Name)
		}
	}
	if migrations[0].Name != "001_create_schema_migrations.sql" || migrations[0].SQL == "" {
		t.Fatalf("first migration = %+v, want schema migration with SQL", migrations[0])
	}
}
