package dbunit

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mysqlDSN    = "root:123456@tcp(127.0.0.1:3306)/"
	postgresDSN = "postgres://postgres:123456@localhost:5432/?sslmode=disable"
)

func checkPostgres(t *testing.T) {
	db, err := sql.Open("pgx", postgresDSN)
	if err != nil {
		t.Skip("skipping postgres tests:", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skip("skipping postgres tests:", err)
	}
}

func checkMySQL(t *testing.T) {
	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		t.Skip("skipping mysql tests:", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skip("skipping mysql tests:", err)
	}
}

func TestNewDatabase(t *testing.T) {
	tests := []struct {
		name   string
		driver string
		dsn    string
		schema string
		check  func(t *testing.T)
	}{
		{
			name:   "mysql",
			driver: "mysql",
			dsn:    mysqlDSN,
			schema: "./testdata/schema-mysql.sql",
			check:  checkMySQL,
		},
		{
			name:   "postgres",
			driver: "pgx",
			dsn:    postgresDSN,
			schema: "./testdata/schema-postgres.sql",
			check:  checkPostgres,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
			tdb := newDatabase(tt.dsn, tt.schema)
			t.Cleanup(func() {
				tdb.Drop()
			})

			db, err := sql.Open(tt.driver, tdb.DSN())
			assert.NoError(t, err)
			defer db.Close()
			assert.NoError(t, db.Ping())
		})
	}
}
