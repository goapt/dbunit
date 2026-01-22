package fixtures

import (
	"database/sql"
	"time"
)

type Helper interface {
	quoteKeyword(str string) string
	databaseName(q *sql.DB) (string, error)
	disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) error
	buildInsertSQL(tableName string, columns []string, records []map[string]any, location *time.Location) (insertSQL, error)
}
