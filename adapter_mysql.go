package dbunit

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLAdapter struct{}

var createTableRegex = regexp.MustCompile(`(?isU)CREATE TABLE\s+.*;`)

func (a *MySQLAdapter) Open(dsn string) (*sql.DB, error) {
	return sql.Open("mysql", dsn)
}

func (a *MySQLAdapter) CreateDatabase(db *sql.DB, name string) error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", name)
	defaultLog.Print(query)
	_, err := db.Exec(query)
	return err
}

func (a *MySQLAdapter) DropDatabase(db *sql.DB, name string) error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)
	defaultLog.Print(query)
	_, err := db.Exec(query)
	return err
}

func (a *MySQLAdapter) ImportSchema(db *sql.DB, schemaFile string) error {
	if !isExists(schemaFile) {
		return fmt.Errorf("sql file not found:%s", schemaFile)
	}

	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	querys := createTableRegex.FindAllString(string(content), -1)

	defaultLog.Print(fmt.Sprintf("Import schema:%s", schemaFile))
	for _, query := range querys {
		defaultLog.Debug(query)
		if len(query) > 0 {
			_, err := db.Exec(query)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *MySQLAdapter) Quote(str string) string {
	return fmt.Sprintf("`%s`", str)
}

func (a *MySQLAdapter) DSN(baseDSN, dbName string) string {
	return baseDSN + dbName + "?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
}

func (a *MySQLAdapter) DriverName() string {
	return "mysql"
}
