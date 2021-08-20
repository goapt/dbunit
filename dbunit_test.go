package dbunit

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestRun(t *testing.T) {
	t.Run("default fixtures", func(t *testing.T) {
		Run(t, "testdata/schema.sql", func(t *testing.T, db *sql.DB) {
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
		Run(t, "testdata/schema.sql", func(t *testing.T, db *sql.DB) {
			row := db.QueryRow("select email from users where id = 1")
			var email string
			if err := row.Scan(&email); err != sql.ErrNoRows {
				t.Fatal(err)
			}
		}, "testdata/fixtures/members.yml", "testdata/fixtures/documents.yml")
	})

	t.Run("custom fixtures", func(t *testing.T) {
		Run(t, "testdata/schema.sql", func(t *testing.T, db *sql.DB) {
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
}

func TestNew(t *testing.T) {
	New(t, func(d *DBUnit) {
		db := d.NewDatabase("testdata/schema.sql", "testdata/fixtures/users.yml")
		// more database
		_ = d.NewDatabase("testdata/schema.sql")
		row := db.QueryRow("select email from users where id = 1")
		var email string
		if err := row.Scan(&email); err != nil {
			t.Fatal(err)
		}

		if email != "test@test.cn" {
			t.Fatalf("user mismatch want %s,but get %s", "test@test.cn", email)
		}
	})
}

func TestLoad(t *testing.T) {
	// SetDatabase("root:123456@tcp(127.0.0.1:33306)/")
	test := NewTest("testdata/schema.sql")
	t.Cleanup(func() {
		test.Drop()
	})

	test.Load("testdata/custom")
}
