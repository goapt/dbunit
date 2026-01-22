package fixtures

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type PostgreSQL struct {
}

func (*PostgreSQL) quoteKeyword(str string) string {
	return fmt.Sprintf(`"%s"`, str)
}

func (*PostgreSQL) databaseName(q *sql.DB) (string, error) {
	var dbName string
	err := q.QueryRow("SELECT current_database()").Scan(&dbName)
	return dbName, err
}

func (h *PostgreSQL) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Requires superuser or replication role privileges
	if _, err = tx.Exec("SET session_replication_role = 'replica';"); err != nil {
		return err
	}

	err = loadFn(tx)

	if _, err2 := tx.Exec("SET session_replication_role = 'origin';"); err2 != nil {
		return err2
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (h *PostgreSQL) buildInsertSQL(tableName string, columns []string, records []map[string]any, location *time.Location) (insertSQL, error) {
	sqlColumnsQuote := make([]string, 0)
	for _, col := range columns {
		sqlColumnsQuote = append(sqlColumnsQuote, h.quoteKeyword(col))
	}

	var sqlBinds []string
	paramCount := 1

	for range records {
		var rowBinds []string
		for range columns {
			rowBinds = append(rowBinds, fmt.Sprintf("$%d", paramCount))
			paramCount++
		}
		sqlBinds = append(sqlBinds, fmt.Sprintf("(%s)", strings.Join(rowBinds, ", ")))
	}

	sqlStr := fmt.Sprintf(
		"INSERT INTO %s(%s) VALUES %s",
		h.quoteKeyword(tableName),
		strings.Join(sqlColumnsQuote, ", "),
		strings.Join(sqlBinds, ", "),
	)

	sqlValues := make([]any, 0)
	for _, record := range records {
		for _, col := range columns {
			val := record[col]
			switch v := val.(type) {
			case string:
				if t, err := tryStrToDate(location, v); err == nil {
					val = t
				}
			case []any, map[any]any:
				val = recursiveToJSON(v)
			}
			sqlValues = append(sqlValues, val)
		}
	}

	return insertSQL{sqlStr, sqlValues}, nil
}
