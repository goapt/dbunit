package fixtures

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

var (
	testDSN  = "root:123456@tcp(127.0.0.1:3306)/"
	mysqlDSN = ""
)

func init() {
	if os.Getenv("DRONE") == "true" {
		testDSN = "root:123456@tcp(database:3306)/"
	}

	if os.Getenv("CI") == "true" {
		testDSN = "root:123456@tcp(127.0.0.1:3306)/"
	}

	mysqlDSN = fmt.Sprintf("%smysql", testDSN)
}

func Test_mySQL_quoteKeyword(t *testing.T) {
	m := &MySQL{}
	k := m.quoteKeyword("status")
	assert.Equal(t, "`status`", k)
}

func Test_mySQL_databaseName(t *testing.T) {
	db, err := sql.Open("mysql", mysqlDSN)
	assert.NoError(t, err)
	m := &MySQL{}
	assert.NoError(t, err)
	s, err := m.databaseName(db)
	assert.NoError(t, err)
	assert.Equal(t, "mysql", s)
}

func Test_mySQL_disableReferentialIntegrity(t *testing.T) {
	db, err := sql.Open("mysql", mysqlDSN)
	assert.NoError(t, err)
	m := &MySQL{}
	err = m.disableReferentialIntegrity(db, func(tx *sql.Tx) error {
		rows, err := tx.Query("select * from user limit 1")
		defer rows.Close()
		if err != nil {
			return err
		}

		return nil
	})
	assert.NoError(t, err)
}
