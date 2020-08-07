package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ilibs/gosql/v2"

	"github.com/goapt/dbunit"
)

func main() {
	dbname := "example"

	db, err := gosql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8&parseTime=True&loc=%s",
		os.Getenv("DEV_DATABASE_USERNAME"),
		os.Getenv("DEV_DATABASE_PASSWORD"),
		os.Getenv("DEV_DATABASE_HOST"), dbname, "Asia%2FShanghai"),
	)
	if err != nil {
		panic(err)
	}

	// 文档数据集
	data, err := dbunit.Dump(db, "testdata/fixtures/documents.yml", "select * from documents limit 3")
	if err != nil {
		panic(err)
	}

	// 用户数据集
	userIds := dbunit.Pluck(data, "user_id")
	_, err = dbunit.Dump(db, "testdata/fixtures/users.yml", "select * from users where id in(?)", userIds)
	if err != nil {
		panic(err)
	}

	// members
	docIds := dbunit.Pluck(data, "doc_id")
	_, err = dbunit.Dump(db, "testdata/fixtures/members.yml", "select * from members where doc_id in(?)", docIds)
	if err != nil {
		panic(err)
	}
}
