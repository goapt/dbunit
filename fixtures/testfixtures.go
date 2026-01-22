package fixtures

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"gopkg.in/yaml.v2"
)

// Loader is the responsible to loading fixtures.
type Loader struct {
	db            *sql.DB
	helper        Helper
	fixturesFiles []*fixtureFile

	skipTestDatabaseCheck bool
	location              *time.Location

	template *Template
}

type insertSQL struct {
	sql    string
	params []any
}

var (
	testDatabaseRegexp    = regexp.MustCompile("(?i)test")
	errDatabaseIsRequired = fmt.Errorf("testfixtures: database is required")
)

// New instantiates a new Loader instance. The "Database" and "Driver"
// options are required.
func New(options ...func(*Loader) error) (*Loader, error) {
	l := &Loader{
		template: NewTemplate(),
		helper:   &MySQL{}, // Default to MySQL
	}

	for _, option := range options {
		if err := option(l); err != nil {
			return nil, err
		}
	}

	if l.db == nil {
		return nil, errDatabaseIsRequired
	}

	if err := l.helper.init(l.db); err != nil {
		return nil, err
	}
	if err := l.buildInsertSQLs(); err != nil {
		return nil, err
	}

	return l, nil
}

// Database sets an existing sql.DB instant to Loader.
func Database(db *sql.DB) func(*Loader) error {
	return func(l *Loader) error {
		l.db = db
		return nil
	}
}

// Dialect sets the database dialect (mysql or postgresql)
func Dialect(dialect string) func(*Loader) error {
	return func(l *Loader) error {
		switch dialect {
		case "postgres", "postgresql", "pgx":
			l.helper = &PostgreSQL{}
		case "mysql":
			l.helper = &MySQL{}
		default:
			return fmt.Errorf("unsupported dialect: %s", dialect)
		}
		return nil
	}
}

// Directory informs Loader to load YAML files from a given directory.
func Directory(dir string) func(*Loader) error {
	return func(l *Loader) error {
		fixtures, err := l.fixturesFromDir(dir)
		if err != nil {
			return err
		}
		l.fixturesFiles = append(l.fixturesFiles, fixtures...)
		return nil
	}
}

// Files informs Loader to load a given set of YAML files.
func Files(files ...string) func(*Loader) error {
	return func(l *Loader) error {
		fixtures, err := l.fixturesFromFiles(files...)
		if err != nil {
			return err
		}
		l.fixturesFiles = append(l.fixturesFiles, fixtures...)
		return nil
	}
}

// Location makes Loader use the given location by default when parsing
// dates. If not given, by default it uses the value of time.Local.
func Location(location *time.Location) func(*Loader) error {
	return func(l *Loader) error {
		l.location = location
		return nil
	}
}

// EnsureTestDatabase returns an error if the database name does not contains
// "test".
func (l *Loader) EnsureTestDatabase() error {
	dbName, err := l.helper.databaseName(l.db)
	if err != nil {
		return err
	}
	if !testDatabaseRegexp.MatchString(dbName) {
		return fmt.Errorf(`testfixtures: database "%s" does not appear to be a test database`, dbName)
	}
	return nil
}

// Load wipes and after load all fixtures in the database.
//
//	if err := fixtures.Load(); err != nil {
//	        ...
//	}
func (l *Loader) Load() error {
	if !l.skipTestDatabaseCheck {
		if err := l.EnsureTestDatabase(); err != nil {
			return err
		}
	}

	// This is MySQL specific, maybe move to helper?
	// But PG doesn't have strict sql_mode.
	// We can ignore error or move to helper.
	_, _ = l.db.Exec("set @@sql_mode=''")

	err := l.helper.disableReferentialIntegrity(l.db, func(tx *sql.Tx) error {
		for _, file := range l.fixturesFiles {
			if _, err := tx.Exec(file.insertSQL.sql, file.insertSQL.params...); err != nil {
				return &InsertError{
					Err:    err,
					File:   file.fileName,
					SQL:    file.insertSQL.sql,
					Params: file.insertSQL.params,
				}
			}
		}
		return nil
	})
	return err
}

// InsertError will be returned if any error happens on database while
// inserting the record.
type InsertError struct {
	Err    error
	File   string
	SQL    string
	Params []any
}

func (e *InsertError) Error() string {
	return fmt.Sprintf(
		"testfixtures: error inserting record: %v, on file: %s, sql: %s, params: %v",
		e.Err,
		e.File,
		e.SQL,
		e.Params,
	)
}

func (l *Loader) buildInsertSQLs() error {
	for _, f := range l.fixturesFiles {
		var records []map[string]any
		if err := yaml.Unmarshal(f.content, &records); err != nil {
			return fmt.Errorf("testfixtures: could not unmarshal YAML: %w", err)
		}

		if len(records) == 0 {
			continue
		}

		var columns []string
		// Get columns from the first record
		for k := range records[0] {
			columns = append(columns, k)
		}
		sort.Strings(columns)

		insertSQL, err := l.helper.buildInsertSQL(f.fileNameWithoutExtension(), columns, records, l.location)
		if err != nil {
			return err
		}
		f.insertSQL = insertSQL
	}

	return nil
}

func (l *Loader) fixturesFromDir(dir string) ([]*fixtureFile, error) {
	fileinfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf(`testfixtures: could not stat directory "%s": %w`, dir, err)
	}

	files := make([]string, 0)

	for _, fileinfo := range fileinfos {
		fileExt := filepath.Ext(fileinfo.Name())
		if !fileinfo.IsDir() && (fileExt == ".yml" || fileExt == ".yaml") {
			files = append(files, path.Join(dir, fileinfo.Name()))
		}
	}
	return l.fixturesFromFiles(files...)
}

func (l *Loader) fixturesFromFiles(fileNames ...string) ([]*fixtureFile, error) {
	var (
		fixtureFiles = make([]*fixtureFile, 0, len(fileNames))
		err          error
	)

	for _, f := range fileNames {
		fixture := &fixtureFile{
			path:     f,
			fileName: filepath.Base(f),
		}
		fixture.content, err = ioutil.ReadFile(fixture.path)
		if err != nil {
			return nil, fmt.Errorf(`testfixtures: could not read file "%s": %w`, fixture.path, err)
		}
		fixture.content, err = l.template.Parse(fixture.content)
		if err != nil {
			return nil, fmt.Errorf(`textfixtures: error on parsing template in %s: %w`, fixture.fileName, err)
		}
		fixtureFiles = append(fixtureFiles, fixture)
	}

	return fixtureFiles, nil
}
