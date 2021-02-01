package fixtures

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixtureFile(t *testing.T) {
	f := &fixtureFile{fileName: "posts.yml"}
	file := f.fileNameWithoutExtension()
	if file != "posts" {
		t.Errorf("Should be 'posts', but returned %s", file)
	}
}

func TestRequiredOptions(t *testing.T) {
	t.Run("DatabaseIsRequired", func(t *testing.T) {
		_, err := New()
		if err != errDatabaseIsRequired {
			t.Error("should return an error if database if not given")
		}
	})
}

const schema = `CREATE TABLE users (
  id int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  user_name varchar(50) NOT NULL DEFAULT '' COMMENT '用户名，用于展示',
  email varchar(100) NOT NULL DEFAULT '' COMMENT '邮箱',
  real_name varchar(50) NOT NULL DEFAULT '' COMMENT '真实姓名',
  password varchar(64) NOT NULL DEFAULT '' COMMENT '密码',
  avatar varchar(100) NOT NULL DEFAULT '' COMMENT '用户头像',
  status int(11) NOT NULL DEFAULT '1' COMMENT '1 启用 2停用',
  about varchar(255) NOT NULL DEFAULT '' COMMENT '个人简介',
  role varchar(30) NOT NULL DEFAULT 'user' COMMENT '用户角色admin,leader,user',
  organization varchar(50) NOT NULL DEFAULT '' COMMENT '部门组织',
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY un_email (email),
  UNIQUE KEY un_user_name (user_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

func TestLoader_Load(t *testing.T) {
	db, err := sql.Open("mysql", testDSN)
	require.NoError(t, err)
	_, err = db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %s`, "testfixtures"))
	require.NoError(t, err)
	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, "testfixtures"))
	require.NoError(t, err)

	db2, err := sql.Open("mysql", fmt.Sprintf("%stestfixtures", testDSN))
	require.NoError(t, err)
	_, err = db2.Exec(schema)
	require.NoError(t, err)
	options := make([]func(*Loader) error, 0)
	options = append(options, Database(db2)) // You database connection

	fs := make([]string, 0)
	fs = append(fs, "../testdata/fixtures/users.yml")
	options = append(options, Files(fs...)) // Specifies the load data file

	f, err := New(options...)
	require.NoError(t, err)
	err = f.Load()
	require.NoError(t, err)

	row := db2.QueryRow("select email from users where id = 1")
	var content string
	err = row.Scan(&content)
	require.NoError(t, err)
	require.Equal(t, `test@test.cn`, content)
}
