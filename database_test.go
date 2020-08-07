package dbunit

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestNewDatabase(t *testing.T) {
	tdb := newDatabase("./testdata/schema.sql")
	t.Cleanup(func() {
		tdb.Drop()
	})

	db, err := sql.Open("mysql", tdb.DSN())
	assert.NoError(t, err)
	defer db.Close()
}
