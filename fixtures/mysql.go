package fixtures

import (
	"database/sql"
	"fmt"
)

type loadFunction func(tx *sql.Tx) error

type mySQL struct {
	tables []string
}

func (h *mySQL) init(db *sql.DB) error {
	var err error
	h.tables, err = h.tableNames(db)
	if err != nil {
		return err
	}

	return nil
}

func (*mySQL) quoteKeyword(str string) string {
	return fmt.Sprintf("`%s`", str)
}

func (*mySQL) databaseName(q *sql.DB) (string, error) {
	var dbName string
	err := q.QueryRow("SELECT DATABASE()").Scan(&dbName)
	return dbName, err
}

func (h *mySQL) tableNames(q *sql.DB) ([]string, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_type = 'BASE TABLE';
	`
	dbName, err := h.databaseName(q)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(query, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil

}

func (h *mySQL) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
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
