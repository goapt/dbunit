package dbunit

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/goapt/dbunit/fixtures"
)

type Testing struct {
	tdb    *database
	db     *sql.DB
	schema string
}

func NewTest(schema string) *Testing {
	tdb := newDatabase(schema)

	// Open connection to the test database.
	// Do NOT import fixtures in a production database!
	// Existing data would be deleted.
	db, err := sql.Open("mysql", tdb.DSN())
	if err != nil {
		panic("test mysql open fail " + err.Error())
	}

	return &Testing{
		tdb,
		db,
		schema,
	}
}

func (d *Testing) DB() *sql.DB {
	return d.db
}

func (d *Testing) Schema() string {
	return d.schema
}

func (d *Testing) Drop() {
	err := d.tdb.Drop()
	if err != nil {
		panic("drop database error " + err.Error())
	}
}

func (d *Testing) Load(files ...string) {
	options := make([]func(*fixtures.Loader) error, 0)
	options = append(options, fixtures.Database(d.db)) // You database connection

	fs := make([]string, 0)
	for _, file := range files {
		if isDir(file) {
			options = append(options, fixtures.Directory(file)) // the directory containing the YAML files
		} else {
			fs = append(fs, file)
		}
	}

	if len(fs) > 0 {
		options = append(options, fixtures.Files(fs...)) // Specifies the load data file
	}

	f, err := fixtures.New(options...)
	if err != nil {
		panic(err)
	}

	fmt.Printf("🐳 Load fixtures:%s\n", strings.Join(files, ","))

	if err := f.Load(); err != nil {
		panic(err)
	}
}

// isDir determines whether the specified path is a directory.
func isDir(path string) bool {
	fio, err := os.Lstat(path)
	if nil != err {
		return false
	}

	return fio.IsDir()
}
