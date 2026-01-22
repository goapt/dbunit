package dbunit

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v2"

	"github.com/goapt/dbunit/fixtures"
)

func parseTableName(query string) string {
	query = strings.ToLower(query)

	// 先找到第一个from位置
	fromIdex := strings.Index(query, "from")
	if fromIdex == -1 {
		return ""
	}

	query = query[fromIdex+4:]

	// 查看有没有as
	asIndex := strings.Index(query, " as")

	if asIndex != -1 {
		return strings.Trim(strings.TrimSpace(query[:asIndex]), "`")
	}

	// 查看有没有where
	whereIndex := strings.Index(query, " where")

	if whereIndex != -1 {
		query = query[:whereIndex]
	}
	s := strings.Split(strings.TrimSpace(query), " ")
	return strings.Trim(strings.TrimSpace(s[0]), "`")
}

func getPrimaryKey(db *sql.DB, driver DriverName, query string) (string, error) {
	tableName := parseTableName(query)

	if tableName == "" {
		return "", errors.New("sql parse table name is empty")
	}

	var pk string

	if driver == MySQLDriver {
		var dbname string
		if err := db.QueryRow("select database()").Scan(&dbname); err != nil {
			return "", err
		}

		row := db.QueryRow("select column_name from information_schema.key_column_usage where constraint_name = 'PRIMARY' and table_schema = ? and table_name = ?", dbname, tableName)
		err := row.Scan(&pk)
		if err != nil {
			return "", err
		}

		return pk, nil
	}

	if driver == PostgresDriver {
		// Postgres primary key query
		// Using information_schema to be consistent, but need to handle schema (defaulting to public)
		// Or assume public.
		// A more robust way in postgres:
		q := `
			SELECT kcu.column_name
			FROM information_schema.key_column_usage kcu
			JOIN information_schema.table_constraints tc
			  ON kcu.constraint_name = tc.constraint_name
			  AND kcu.table_schema = tc.table_schema
			WHERE kcu.table_name = $1
			  AND tc.constraint_type = 'PRIMARY KEY'
			LIMIT 1
		`
		row := db.QueryRow(q, tableName)
		err := row.Scan(&pk)
		if err != nil {
			return "", err
		}
		return pk, nil
	}

	return "", fmt.Errorf("unsupported database or failed to get primary key")
}

func Dump(db *sql.DB, filePath, query string, args ...any) ([]map[string]any, error) {
	// Check if Postgres to convert placeholders
	var isPostgres bool
	var dbname string
	var driver = MySQLDriver
	if err := db.QueryRow("select current_database()").Scan(&dbname); err == nil {
		isPostgres = true
		driver = PostgresDriver
	}

	pk, err := getPrimaryKey(db, driver, query)
	if err != nil {
		return nil, fmt.Errorf("get primary key error %w", err)
	}

	query, newArgs, err := inReplace(query, args...)
	if err != nil {
		return nil, err
	}

	if isPostgres {
		query = convertToPostgresPlaceholders(query)
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(newArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var oldData = make([]map[string]any, 0)
	if isExists(filePath) {
		d, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		tpl := fixtures.NewTemplate()
		d, err = tpl.Parse(d)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(d, &oldData)
		if err != nil {
			return nil, err
		}
	}

	fixturesSlice := make([]yaml.MapSlice, 0, 10)
	fixtureMaps := make([]map[string]any, 0)
	for rows.Next() {
		entries := make([]any, len(columns))
		entryPtrs := make([]any, len(entries))
		for i := range entries {
			entryPtrs[i] = &entries[i]
		}
		if err := rows.Scan(entryPtrs...); err != nil {
			return nil, err
		}

		entryMap := make([]yaml.MapItem, len(entries))
		entryMap2 := make(map[string]any)
		for i, column := range columns {
			v := convertValue(entries[i])
			entryMap[i] = yaml.MapItem{
				Key:   column,
				Value: v,
			}
			entryMap2[column] = v
		}

		if !isDuplicate(oldData, entryMap2, pk) {
			fixturesSlice = append(fixturesSlice, entryMap)
		} else {
			if pk != "" {
				fmt.Println(fmt.Sprintf("[duplicate] %s ignore primary key:%v", filePath, entryMap2[pk]))
			} else {
				fmt.Println(fmt.Sprintf("[duplicate] %s ignore primary key", filePath))
			}
		}
		fixtureMaps = append(fixtureMaps, entryMap2)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(fixturesSlice) == 0 && len(oldData) != 0 {
		return fixtureMaps, nil
	}

	err = writeYml(filePath, fixturesSlice, len(oldData))
	return fixtureMaps, err
}

func convertToPostgresPlaceholders(query string) string {
	// Simple replacement of ? with $1, $2, etc.
	// This assumes that ? are only used as placeholders and not in strings.
	// A robust parser is needed for perfect safety, but simple replacement is common in simple helpers.
	count := 0
	var buf strings.Builder
	for _, char := range query {
		if char == '?' {
			count++
			buf.WriteString(fmt.Sprintf("$%d", count))
		} else {
			buf.WriteRune(char)
		}
	}
	return buf.String()
}

func writeYml(filePath string, fixtures []yaml.MapSlice, oldlen int) error {
	var f *os.File
	var err error
	if isExists(filePath) && oldlen != 0 {
		f, err = os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0666)
	} else {
		f, err = os.Create(filePath)
	}

	if err != nil {
		return err
	}
	defer f.Close()

	data, err := yaml.Marshal(fixtures)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte("\n#Created At:" + time.Now().Format("2006-01-02 13:04:05") + "\n"))
	_, err = f.Write(data)
	return err
}

func convertValue(value any) any {
	switch v := value.(type) {
	case []byte:
		if utf8.Valid(v) {
			return string(v)
		}
	}
	return value
}

func isDuplicate(x []map[string]any, y map[string]any, pk string) bool {
	for _, v := range x {
		if pk != "" {
			v1 := fmt.Sprintf("%v", v[pk])
			v2 := fmt.Sprintf("%v", y[pk])
			if v1 == v2 {
				return true
			}
		} else {
			v1 := fmt.Sprintf("%v", filterDate(v))
			v2 := fmt.Sprintf("%v", filterDate(y))
			if v1 == v2 {
				return true
			}
		}
	}

	return false
}

func filterDate(m map[string]any) map[string]any {
	newMap := make(map[string]any)
	for k, v := range m {
		if !tryParseDate(fmt.Sprintf("%s", v)) {
			newMap[k] = v
		}
	}
	return newMap
}

var timeFormats = [...]string{
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02 15:04:05 Z0700 CST",
	"2006-01-02 15:04:05Z0700",
	"2006-01-02 15:04:05 MST",
}

func tryParseDate(s string) bool {
	for _, f := range timeFormats {
		_, err := time.Parse(f, s)
		if err == nil {
			return true
		}
	}
	return false
}

func asSliceForIn(i any) (v reflect.Value, ok bool) {
	if i == nil {
		return reflect.Value{}, false
	}

	v = reflect.ValueOf(i)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Only expand slices
	if t.Kind() != reflect.Slice {
		return reflect.Value{}, false
	}

	// []byte is a driver.Value type so it should not be expanded
	if t == reflect.TypeOf([]byte{}) {
		return reflect.Value{}, false

	}

	return v, true
}

// inReplace expands slice values in args, returning the modified query string
// and a new arg list that can be executed by a database. The `query` should
// use the `?` bindVar.  The return value uses the `?` bindVar.
func inReplace(query string, args ...any) (string, []any, error) {
	// argMeta stores reflect.Value and length for slices and
	// the value itself for non-slice arguments
	type argMeta struct {
		v      reflect.Value
		i      any
		length int
	}

	var flatArgsCount int
	var anySlices bool

	var stackMeta [32]argMeta

	var meta []argMeta
	if len(args) <= len(stackMeta) {
		meta = stackMeta[:len(args)]
	} else {
		meta = make([]argMeta, len(args))
	}

	for i, arg := range args {
		if a, ok := arg.(driver.Valuer); ok {
			var err error
			arg, err = a.Value()
			if err != nil {
				return "", nil, err
			}
		}

		if v, ok := asSliceForIn(arg); ok {
			meta[i].length = v.Len()
			meta[i].v = v

			anySlices = true
			flatArgsCount += meta[i].length

			if meta[i].length == 0 {
				return "", nil, errors.New("empty slice passed to 'in' query")
			}
		} else {
			meta[i].i = arg
			flatArgsCount++
		}
	}

	// don't do any parsing if there aren't any slices;  note that this means
	// some errors that we might have caught below will not be returned.
	if !anySlices {
		return query, args, nil
	}

	newArgs := make([]any, 0, flatArgsCount)

	var buf strings.Builder
	buf.Grow(len(query) + len(", ?")*flatArgsCount)

	var arg, offset int

	for i := strings.IndexByte(query[offset:], '?'); i != -1; i = strings.IndexByte(query[offset:], '?') {
		if arg >= len(meta) {
			// if an argument wasn't passed, lets return an error;  this is
			// not actually how database/sql Exec/Query works, but since we are
			// creating an argument list programmatically, we want to be able
			// to catch these programmer errors earlier.
			return "", nil, errors.New("number of bindVars exceeds arguments")
		}

		argMeta := meta[arg]
		arg++

		// not a slice, continue.
		// our questionmark will either be written before the next expansion
		// of a slice or after the loop when writing the rest of the query
		if argMeta.length == 0 {
			offset = offset + i + 1
			newArgs = append(newArgs, argMeta.i)
			continue
		}

		// write everything up to and including our ? character
		buf.WriteString(query[:offset+i+1])

		for si := 1; si < argMeta.length; si++ {
			buf.WriteString(", ?")
		}

		newArgs = appendReflectSlice(newArgs, argMeta.v, argMeta.length)

		// slice the query and reset the offset. this avoids some bookkeeping for
		// the write after the loop
		query = query[offset+i+1:]
		offset = 0
	}

	buf.WriteString(query)

	if arg < len(meta) {
		return "", nil, errors.New("number of bindVars less than number arguments")
	}

	return buf.String(), newArgs, nil
}

func appendReflectSlice(args []any, v reflect.Value, vlen int) []any {
	switch val := v.Interface().(type) {
	case []any:
		args = append(args, val...)
	case []int:
		for i := range val {
			args = append(args, val[i])
		}
	case []string:
		for i := range val {
			args = append(args, val[i])
		}
	default:
		for si := 0; si < vlen; si++ {
			args = append(args, v.Index(si).Interface())
		}
	}

	return args
}
