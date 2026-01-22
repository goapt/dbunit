package fixtures

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type loadFunction func(tx *sql.Tx) error

type MySQL struct {
}

func (*MySQL) quoteKeyword(str string) string {
	return fmt.Sprintf("`%s`", str)
}

func (*MySQL) databaseName(q *sql.DB) (string, error) {
	var dbName string
	err := q.QueryRow("SELECT DATABASE()").Scan(&dbName)
	return dbName, err
}

func (h *MySQL) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return err
	}

	err = loadFn(tx)
	_, err2 := tx.Exec("SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}

	return tx.Commit()
}

func (h *MySQL) buildInsertSQL(tableName string, columns []string, records []map[string]any, location *time.Location) (insertSQL, error) {
	sqlColumnsQuote := make([]string, 0)
	sqlValuesBind := make([]string, 0)

	// Quote columns and prepare placeholders
	for _, col := range columns {
		sqlColumnsQuote = append(sqlColumnsQuote, h.quoteKeyword(col))
		sqlValuesBind = append(sqlValuesBind, "?")
	}

	sqlBind := fmt.Sprintf("(%s)", strings.Join(sqlValuesBind, ", "))
	sqlBinds := make([]string, len(records))
	for k := range records {
		sqlBinds[k] = sqlBind
	}

	sqlStr := fmt.Sprintf(
		"REPLACE INTO %s(%s) VALUES %s",
		h.quoteKeyword(tableName),
		strings.Join(sqlColumnsQuote, ", "),
		strings.Join(sqlBinds, ", "),
	)

	sqlValues := make([]any, 0)

	for _, record := range records {
		for _, col := range columns {
			// Extract value using unquoted column name
			// The original logic in testfixtures.go used:
			// k = strings.Trim(k, "`") which implies it expected quoted keys or just stripped them just in case.
			// But records map keys are usually unquoted.
			// Wait, testfixtures.go:182: for k, _ := range records[0] { ... }
			// Then testfixtures.go:203: for _, k := range sqlColumnsQuote { k = strings.Trim(k, "`") ... }
			// This was a bit convoluted. Since we pass `columns` (which should be raw names), we can just use them.

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
