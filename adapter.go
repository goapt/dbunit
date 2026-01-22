package dbunit

import "database/sql"

type DBAdapter interface {
	Open(dsn string) (*sql.DB, error)
	CreateDatabase(db *sql.DB, name string) error
	DropDatabase(db *sql.DB, name string) error
	ImportSchema(db *sql.DB, schemaFile string) error
	Quote(str string) string
	DSN(baseDSN, dbName string) string
	DriverName() string
}
