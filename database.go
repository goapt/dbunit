package dbunit

import (
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type DriverName string

var (
	id             atomic.Int32
	MySQLDriver    DriverName = "mysql"
	PostgresDriver DriverName = "postgres"
)

type database struct {
	Name    string
	source  string
	db      *sql.DB
	adapter DBAdapter
}

func newDatabase(dsn string, schema string) *database {
	newID := id.Add(1)
	name := fmt.Sprintf("test_%d_%d", time.Now().UnixNano(), newID)
	return newDatabaseWithName(dsn, name, schema)
}

func newDatabaseWithName(dsn string, name string, schema string) *database {
	var adapter DBAdapter
	if strings.HasPrefix(dsn, string(PostgresDriver)) {
		adapter = &PostgresAdapter{}
	} else {
		adapter = &MySQLAdapter{}
	}

	db := &database{
		Name:    name,
		source:  dsn,
		adapter: adapter,
	}
	err := db.connection()

	if err != nil {
		panic("test database connection fail," + err.Error())
	}

	err = db.create()
	if err != nil {
		panic("test database create database fail," + err.Error())
	}

	err = db.Import(schema)
	if err != nil {
		panic(err)
	}
	return db
}

func (d *database) DSN() string {
	return d.adapter.DSN(d.source, d.Name)
}

func (d *database) connection() error {
	db, err := d.adapter.Open(d.source)
	if err != nil {
		return err
	}
	d.db = db
	return nil
}

func (d *database) Drop() error {
	err := d.adapter.DropDatabase(d.db, d.Name)
	if err != nil {
		return err
	}
	d.db.Close()
	return nil
}

func (d *database) create() error {
	return d.adapter.CreateDatabase(d.db, d.Name)
}

func (d *database) Import(schema string) error {
	// Import schema needs to connect to the new database
	db, err := d.adapter.Open(d.DSN())
	if err != nil {
		return err
	}
	defer db.Close()

	return d.adapter.ImportSchema(db, schema)
}
