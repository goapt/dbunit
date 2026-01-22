package dbunit

import (
	"database/sql"
	"testing"
)

func TestDumpSQL(t *testing.T) {
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
			Run(t, tt.dsn, tt.schema, func(t *testing.T, db *sql.DB) {
				data, err := Dump(db, "testdata/fixtures/documents.yml", "select * from documents limit 10")
				if err != nil {
					t.Fatal("dump documents error:", err)
				}

				userIds := Pluck(data, "user_id")

				_, err = Dump(db, "testdata/fixtures/users.yml", "select * from users where id in(?)", userIds)
				if err != nil {
					t.Fatal("dump users error:", err)
				}

				_, err = Dump(db, "testdata/fixtures/members.yml", "select * from members where id = 0")
				if err != nil {
					t.Fatal("dump users error:", err)
				}
			})
		})
	}
}

func Test_getPrimaryKey(t *testing.T) {
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
			Run(t, tt.dsn, tt.schema, func(t *testing.T, db *sql.DB) {
				pk, err := getPrimaryKey(db, tt.driver, "select * from users limit 1")

				if err != nil {
					t.Fatal("getPrimaryKey error", err)
				}

				if pk != "id" {
					t.Fatal("getPrimaryKey error must get id")
				}
			})
		})
	}

}
