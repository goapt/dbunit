package dbunit

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresSupport(t *testing.T) {
	// User provided DSN
	pgDSN := "postgresql://postgres:123456@localhost:5432/?sslmode=disable"

	// Check connectivity first
	tempDB, err := sql.Open("pgx", pgDSN)
	if err != nil {
		t.Fatalf("Skipping PostgreSQL test: failed to open driver: %v", err)
	}
	if err := tempDB.Ping(); err != nil {
		t.Fatalf("Skipping PostgreSQL test: failed to ping database: %v. Please ensure Postgres is running and DSN %s is correct.", err, pgDSN)
	}
	tempDB.Close()

	// Save original DSN
	originalDSN := defaultTestDSN
	// Set Postgres DSN
	SetDatabase(pgDSN)
	defer SetDatabase(originalDSN)

	// 1. Test Database Creation and Schema Import
	// newDatabase might panic if it fails, so we can defer recover if we want to be safe,
	// but since we checked Ping, it should work unless permissions issue for CREATE DATABASE.
	tdb := newDatabase("./testdata/schema-postgres.sql")
	// Ensure cleanup
	defer func() {
		_ = tdb.Drop()
	}()

	// Verify connection
	db, err := tdb.adapter.Open(tdb.DSN())
	if err != nil {
		t.Fatalf("Failed to open postgres connection: %v", err)
	}
	defer db.Close()

	// Verify table created
	var tableName string
	err = db.QueryRow("SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename = 'users'").Scan(&tableName)
	assert.NoError(t, err, "Table users should exist")
	assert.Equal(t, "users", tableName)

	// 2. Test Fixture Loading
	// Create Testing instance manually to simulate NewTest behavior but with existing tdb
	testingDB := &Testing{
		tdb:    tdb,
		db:     db,
		schema: "./testdata/schema-postgres.sql",
	}

	// Load fixtures
	// We need to pass the dialect option manually since Testing.Load does it internally based on tdb.
	// Actually Testing.Load uses d.tdb.adapter.DriverName(), which is set on tdb.
	// So calling testingDB.Load should work fine.
	testingDB.Load("./testdata/fixtures/users.yml")

	// Check data
	var count int
	err = db.QueryRow("SELECT count(*) FROM users").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count, "Should have 2 users loaded")

	var userName string
	err = db.QueryRow("SELECT user_name FROM users WHERE id = 1").Scan(&userName)
	assert.NoError(t, err)
	assert.Equal(t, "test1", userName)
}
