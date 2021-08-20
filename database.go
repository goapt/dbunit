package dbunit

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sync/atomic"
	"time"
)

var (
	defaultTestDSN         = "root:123456@tcp(127.0.0.1:3306)/"
	createTableRegex       = regexp.MustCompile(`(?isU)CREATE TABLE\s+.*;`)
	id               int32 = 0
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
	Name   string
	source string
	db     *sql.DB
}

func newDatabase(schema string) *database {
	atomic.AddInt32(&id, 1)
	name := "test_" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), id)
	return newDatabaseWithName(name, schema)
}

func newDatabaseWithName(name string, schema string) *database {
	db := &database{Name: name, source: defaultTestDSN}
	err := db.connection()

	if err != nil {
		panic("test mysql connection fail," + err.Error())
	}

	err = db.create()
	if err != nil {
		panic("test mysql create database fail," + err.Error())
	}

	err = db.Import(schema)
	if err != nil {
		panic(err)
	}
	return db
}

func (d *database) DSN() string {
	return defaultTestDSN + d.Name + "?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
}

func (d *database) connection() error {
	db, err := sql.Open("mysql", d.source)
	if err != nil {
		return err
	}
	d.db = db
	return nil
}

func (d *database) Drop() error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", d.Name)
	defaultLog.Print(query)
	_, err := d.db.Exec(query)
	if err != nil {
		return err
	}
	d.db.Close()
	return nil
}

func (d *database) create() error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", d.Name)
	defaultLog.Print(query)
	_, err := d.db.Exec(query)
	return err
}

func (d *database) Import(schema string) error {

	if !isExists(schema) {
		return fmt.Errorf("sql file not found:%s", schema)
	}

	content, err := ioutil.ReadFile(schema)
	if err != nil {
		return err
	}

	querys := createTableRegex.FindAllString(string(content), -1)

	var results []sql.Result

	db, err := sql.Open("mysql", d.DSN())
	if err != nil {
		return err
	}

	defaultLog.Print(fmt.Sprintf("Import schema:%s", schema))
	for _, query := range querys {
		defaultLog.Debug(query)
		if len(query) > 0 {
			result, err := db.Exec(query)
			results = append(results, result)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
