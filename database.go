package dbunit

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var (
	defaultTestDSN       = "root:123456@tcp(127.0.0.1:3306)/"
	id             int32 = 0
)

func init() {
	if os.Getenv("DRONE") == "true" {
		SetDatabase("root:123456@tcp(database:3306)/")
	}

	if os.Getenv("CI") == "true" {
		SetDatabase("root:root@tcp(127.0.0.1:3306)/")
	}
}

// SetDatabase 配置单元测试的数据库DSN
func SetDatabase(dsn string) {
	defaultTestDSN = dsn
}

type database struct {
	Name    string
	source  string
	db      *sql.DB
	adapter DBAdapter
}

func newDatabase(schema string) *database {
	atomic.AddInt32(&id, 1)
	name := "test_" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), id)
	return newDatabaseWithName(name, schema)
}

func newDatabaseWithName(name string, schema string) *database {
	var adapter DBAdapter
	if strings.HasPrefix(defaultTestDSN, "postgres") {
		adapter = &PostgresAdapter{}
	} else {
		adapter = &MySQLAdapter{}
	}

	db := &database{
		Name:    name,
		source:  defaultTestDSN,
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
	return d.adapter.DSN(defaultTestDSN, d.Name)
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
