package dbunit

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresAdapter struct{}

func (a *PostgresAdapter) Open(dsn string) (*sql.DB, error) {
	return sql.Open("pgx", dsn)
}

func (a *PostgresAdapter) CreateDatabase(db *sql.DB, name string) error {
	query := fmt.Sprintf(`CREATE DATABASE "%s"`, name)
	defaultLog.Print(query)
	_, err := db.Exec(query)
	return err
}

func (a *PostgresAdapter) DropDatabase(db *sql.DB, name string) error {
	// Disconnect users first?
	// For local tests, usually no other users.
	query := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, name)
	defaultLog.Print(query)
	_, err := db.Exec(query)
	return err
}

func (a *PostgresAdapter) ImportSchema(db *sql.DB, schemaFile string) error {
	if !isExists(schemaFile) {
		return fmt.Errorf("sql file not found:%s", schemaFile)
	}

	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	// Simple split by ; and trim
	queries := strings.Split(string(content), ";")

	defaultLog.Print(fmt.Sprintf("Import schema:%s", schemaFile))
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		defaultLog.Debug(query)
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("error executing query: %s, err: %w", query, err)
		}
	}
	return nil
}

func (a *PostgresAdapter) Quote(str string) string {
	return fmt.Sprintf(`"%s"`, str)
}

func (a *PostgresAdapter) DSN(baseDSN, dbName string) string {
	u, err := url.Parse(baseDSN)
	if err != nil {
		// Fallback to simple concatenation if parse fails
		dsn := baseDSN
		if !strings.HasSuffix(dsn, "/") {
			dsn += "/"
		}
		dsn += dbName + "?sslmode=disable"
		return dsn
	}

	u.Path = "/" + dbName
	q := u.Query()
	if q.Get("sslmode") == "" {
		q.Set("sslmode", "disable")
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (a *PostgresAdapter) DriverName() string {
	return "pgx"
}
