package fixtures

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"
)

// Loader is the responsible to loading fixtures.
type Loader struct {
	db            *sql.DB
	helper        *mySQL
	fixturesFiles []*fixtureFile

	skipTestDatabaseCheck bool
	location              *time.Location

	template           bool
	templateFuncs      template.FuncMap
	templateLeftDelim  string
	templateRightDelim string
	templateOptions    []string
	templateData       interface{}
}

type fixtureFile struct {
	path      string
	fileName  string
	content   []byte
	insertSQL insertSQL
}

type insertSQL struct {
	sql    string
	params []interface{}
}

var (
	testDatabaseRegexp    = regexp.MustCompile("(?i)test")
	errDatabaseIsRequired = fmt.Errorf("testfixtures: database is required")
)

// New instantiates a new Loader instance. The "Database" and "Driver"
// options are required.
func New(options ...func(*Loader) error) (*Loader, error) {
	l := &Loader{
		templateLeftDelim:  "{{",
		templateRightDelim: "}}",
		templateOptions:    []string{"missingkey=zero"},
	}

	for _, option := range options {
		if err := option(l); err != nil {
			return nil, err
		}
	}

	if l.db == nil {
		return nil, errDatabaseIsRequired
	}

	l.helper = &mySQL{}

	if err := l.helper.init(l.db); err != nil {
		return nil, err
	}
	if err := l.buildInsertSQLs(); err != nil {
		return nil, err
	}

	l.templateFuncs = make(template.FuncMap)
	l.templateFuncs["now"] = func() string {
		layout := "2006-01-02 13:04:05"
		return time.Now().Format(layout)
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

// Template makes loader process each YAML file as an template using the
// text/template package.
//
// For more information on how templates work in Go please read:
// https://golang.org/pkg/text/template/
//
// If not given the YAML files are parsed as is.
func Template() func(*Loader) error {
	return func(l *Loader) error {
		l.template = true
		return nil
	}
}

// TemplateFuncs allow choosing which functions will be available
// when processing templates.
//
// For more information see: https://golang.org/pkg/text/template/#Template.Funcs
func TemplateFuncs(funcs template.FuncMap) func(*Loader) error {
	return func(l *Loader) error {
		if !l.template {
			return fmt.Errorf(`testfixtures: the Template() options is required in order to use the TemplateFuns() option`)
		}

		for k, v := range funcs {
			l.templateFuncs[k] = v
		}
		return nil
	}
}

// TemplateDelims allow choosing which delimiters will be used for templating.
// This defaults to "{{" and "}}".
//
// For more information see https://golang.org/pkg/text/template/#Template.Delims
func TemplateDelims(left, right string) func(*Loader) error {
	return func(l *Loader) error {
		if !l.template {
			return fmt.Errorf(`testfixtures: the Template() options is required in order to use the TemplateDelims() option`)
		}

		l.templateLeftDelim = left
		l.templateRightDelim = right
		return nil
	}
}

// TemplateOptions allows you to specific which text/template options will
// be enabled when processing templates.
//
// This defaults to "missingkey=zero". Check the available options here:
// https://golang.org/pkg/text/template/#Template.Option
func TemplateOptions(options ...string) func(*Loader) error {
	return func(l *Loader) error {
		if !l.template {
			return fmt.Errorf(`testfixtures: the Template() options is required in order to use the TemplateOptions() option`)
		}

		l.templateOptions = options
		return nil
	}
}

// TemplateData allows you to specify which data will be available
// when processing templates. Data is accesible by prefixing it with a "."
// like {{.MyKey}}.
func TemplateData(data interface{}) func(*Loader) error {
	return func(l *Loader) error {
		if !l.template {
			return fmt.Errorf(`testfixtures: the Template() options is required in order to use the TemplateData() option`)
		}

		l.templateData = data
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
//     if err := fixtures.Load(); err != nil {
//             ...
//     }
func (l *Loader) Load() error {
	if !l.skipTestDatabaseCheck {
		if err := l.EnsureTestDatabase(); err != nil {
			return err
		}
	}

	l.db.Exec("set @@sql_mode=''")

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
	Params []interface{}
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
		var records []map[string]interface{}
		if err := yaml.Unmarshal(f.content, &records); err != nil {
			return fmt.Errorf("testfixtures: could not unmarshal YAML: %w", err)
		}

		if len(records) == 0 {
			continue
		}

		sqlColumnsQuote := make([]string, 0)
		sqlValuesBind := make([]string, 0)
		for k, _ := range records[0] {
			sqlColumnsQuote = append(sqlColumnsQuote, l.helper.quoteKeyword(k))
			sqlValuesBind = append(sqlValuesBind, "?")
		}
		sqlBind := fmt.Sprintf("(%s)", strings.Join(sqlValuesBind, ", "))
		sqlBinds := make([]string, len(records))
		for k, _ := range records {
			sqlBinds[k] = sqlBind
		}
		sort.Strings(sqlColumnsQuote)
		var (
			sqlStr = fmt.Sprintf(
				"REPLACE INTO %s(%s) VALUES %s",
				l.helper.quoteKeyword(f.fileNameWithoutExtension()),
				strings.Join(sqlColumnsQuote, ", "),
				strings.Join(sqlBinds, ", "),
			)
			sqlValues = make([]interface{}, 0)
		)

		for _, record := range records {
			for _, k := range sqlColumnsQuote {
				k = strings.Trim(k, "`")
				switch v := record[k].(type) {
				case string:
					if t, err := l.tryStrToDate(v); err == nil {
						record[k] = t
					}
				case []interface{}, map[interface{}]interface{}:
					record[k] = recursiveToJSON(v)
				}
				sqlValues = append(sqlValues, record[k])
			}
		}

		f.insertSQL = insertSQL{sqlStr, sqlValues}
	}

	return nil
}

func (f *fixtureFile) fileNameWithoutExtension() string {
	return strings.Replace(f.fileName, filepath.Ext(f.fileName), "", 1)
}

func (l *Loader) fixturesFromDir(dir string) ([]*fixtureFile, error) {
	fileinfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf(`testfixtures: could not stat directory "%s": %w`, dir, err)
	}

	files := make([]*fixtureFile, 0, len(fileinfos))

	for _, fileinfo := range fileinfos {
		fileExt := filepath.Ext(fileinfo.Name())
		if !fileinfo.IsDir() && (fileExt == ".yml" || fileExt == ".yaml") {
			fixture := &fixtureFile{
				path:     path.Join(dir, fileinfo.Name()),
				fileName: fileinfo.Name(),
			}
			fixture.content, err = ioutil.ReadFile(fixture.path)
			if err != nil {
				return nil, fmt.Errorf(`testfixtures: could not read file "%s": %w`, fixture.path, err)
			}
			if err := l.processFileTemplate(fixture); err != nil {
				return nil, err
			}
			files = append(files, fixture)
		}
	}
	return files, nil
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
		if err := l.processFileTemplate(fixture); err != nil {
			return nil, err
		}
		fixtureFiles = append(fixtureFiles, fixture)
	}

	return fixtureFiles, nil
}

func (l *Loader) processFileTemplate(f *fixtureFile) error {
	if !l.template {
		return nil
	}

	t := template.New("").
		Funcs(l.templateFuncs).
		Delims(l.templateLeftDelim, l.templateRightDelim).
		Option(l.templateOptions...)
	t, err := t.Parse(string(f.content))
	if err != nil {
		return fmt.Errorf(`textfixtures: error on parsing template in %s: %w`, f.fileName, err)
	}

	var buffer bytes.Buffer
	if err := t.Execute(&buffer, l.templateData); err != nil {
		return fmt.Errorf(`textfixtures: error on executing template in %s: %w`, f.fileName, err)
	}

	f.content = buffer.Bytes()
	return nil
}
