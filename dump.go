package dbunit

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ilibs/gosql/v2"
	"github.com/jmoiron/sqlx"
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

func getPrimaryKey(db *gosql.DB, query string) (string, error) {
	tableName := parseTableName(query)

	if tableName == "" {
		return "", errors.New("sql parse table name is empty")
	}

	row := db.QueryRowx("select database();")
	var dbname string
	err := row.Scan(&dbname)
	if err != nil {
		return "", fmt.Errorf("get database name error %w", err)
	}

	row = db.QueryRowx("select column_name from information_schema.key_column_usage where constraint_name = 'PRIMARY' and table_schema = ? and table_name = ?", dbname, tableName)
	var pk string
	err = row.Scan(&pk)
	if err != nil {
		return "", err
	}
	return pk, nil
}

func Dump(db *gosql.DB, filePath, query string, args ...interface{}) ([]map[string]interface{}, error) {
	pk, err := getPrimaryKey(db, query)
	if err != nil {
		return nil, fmt.Errorf("get primary key error %w", err)
	}

	query, newArgs, err := sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}

	stmt, err := db.Preparex(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Queryx(newArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var oldData = make([]map[string]interface{}, 0)
	if isExists(filePath) {
		d, err := ioutil.ReadFile(filePath)
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
	fixtureMaps := make([]map[string]interface{}, 0)
	for rows.Next() {
		entries := make([]interface{}, len(columns))
		entryPtrs := make([]interface{}, len(entries))
		for i := range entries {
			entryPtrs[i] = &entries[i]
		}
		if err := rows.Scan(entryPtrs...); err != nil {
			return nil, err
		}

		entryMap := make([]yaml.MapItem, len(entries))
		entryMap2 := make(map[string]interface{})
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
			fmt.Println(fmt.Sprintf("[duplicate] %s ignore primary key:%v", filePath, entryMap2[pk]))
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

func convertValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []byte:
		if utf8.Valid(v) {
			return string(v)
		}
	}
	return value
}

func isDuplicate(x []map[string]interface{}, y map[string]interface{}, pk string) bool {
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

func filterDate(m map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
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
