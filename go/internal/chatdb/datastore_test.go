package chatdb

import (
	"testing"

	"berty.tech/go/internal/chatdb/migrations"
	"berty.tech/go/internal/gormutil"
	"berty.tech/go/pkg/chatmodel"
	"go.uber.org/zap"
)

func TestDropDatabase(t *testing.T) {
	db := TestingSqliteDB(t, zap.NewNop())

	err := DropDatabase(db)
	if err != nil {
		t.Fatalf("DropDatabase failed: %v", err)
	}

	count := len(gormutil.TestingGetTableNames(t, db))
	if count > 0 {
		t.Fatalf("Expected 0 tables, got %d.", count)
	}
}

func TestAllTables(t *testing.T) {
	db := TestingSqliteDB(t, zap.NewNop())
	gormutil.TestingHasExpectedTables(t, db, chatmodel.AllTables())
}

func TestAllMigrations(t *testing.T) {
	migrations := migrations.GetMigrations()
	if len(migrations) == 0 {
		t.Log("No migrations specified")
		t.Skip()
	}

	gormutil.TestingMigrationsVSAutoMigrate(t, migrations, chatmodel.AllModels(), zap.NewNop())
}
