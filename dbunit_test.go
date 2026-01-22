package dbunit

import (
	"database/sql"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name   string
		driver DriverName
		dsn    string
		schema string
		check  func(t *testing.T)
	}{
		{
			name:   "mysql",
			driver: MySQLDriver,
			dsn:    mysqlDSN,
			schema: "testdata/schema-mysql.sql",
			check:  checkMySQL,
		},
		{
			name:   "postgres",
			driver: PostgresDriver,
			dsn:    postgresDSN,
			schema: "testdata/schema-postgres.sql",
			check:  checkPostgres,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
			t.Run("default fixtures", func(t *testing.T) {
				Run(t, tt.dsn, tt.schema, func(t *testing.T, db *sql.DB) {
					row := db.QueryRow("select email from users where id = 1")
					var email string
					if err := row.Scan(&email); err != nil {
						t.Fatal(err)
					}

					if email != "test@test.cn" {
						t.Fatalf("user mismatch want %s,but get %s", "test@test.cn", email)
					}
				})
			})

			t.Run("select fixtures", func(t *testing.T) {
				Run(t, tt.dsn, tt.schema, func(t *testing.T, db *sql.DB) {
					row := db.QueryRow("select email from users where id = 1")
					var email string
					if err := row.Scan(&email); err != sql.ErrNoRows {
						t.Fatal(err)
					}
				}, "testdata/fixtures/members.yml", "testdata/fixtures/documents.yml")
			})

			t.Run("custom fixtures", func(t *testing.T) {
				Run(t, tt.dsn, tt.schema, func(t *testing.T, db *sql.DB) {
					var ct int
					err := db.QueryRow("select count(1) from custom").Scan(&ct)

					if err != nil {
						t.Fatal(err)
					}

					if ct == 0 {
						t.Fatalf("user mismatch want %s,but get %d", " > 0", ct)
					}
				}, "testdata/custom")
			})
		})
	}
}

func TestNew(t *testing.T) {
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
			schema: "testdata/schema-mysql.sql",
			check:  checkMySQL,
		},
		{
			name:   "postgres",
			driver: "pgx",
			dsn:    postgresDSN,
			schema: "testdata/schema-postgres.sql",
			check:  checkPostgres,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
			New(t, func(d *DBUnit) {
				db := d.NewDatabase(tt.dsn, tt.schema, "testdata/fixtures/users.yml")
				// more database
				_ = d.NewDatabase(tt.dsn, tt.schema)
				row := db.QueryRow("select email from users where id = 1")
				var email string
				if err := row.Scan(&email); err != nil {
					t.Fatal(err)
				}

				if email != "test@test.cn" {
					t.Fatalf("user mismatch want %s,but get %s", "test@test.cn", email)
				}
			})
		})
	}
}

func TestLoad(t *testing.T) {
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
			schema: "testdata/schema-mysql.sql",
			check:  checkMySQL,
		},
		{
			name:   "postgres",
			driver: "pgx",
			dsn:    postgresDSN,
			schema: "testdata/schema-postgres.sql",
			check:  checkPostgres,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
			test := NewTest(tt.dsn, tt.schema)
			t.Cleanup(func() {
				test.Drop()
			})

			test.Load("testdata/custom")
		})
	}
}
