package dbunit

import (
	"path/filepath"
	"testing"

	"github.com/ilibs/gosql/v2"
)

func Run(t *testing.T, schema string, f func(t *testing.T, db *gosql.DB), fixtures ...string) {
	New(t, func(d *DBUnit) {
		db := d.NewDatabase(schema, fixtures...)
		f(t, db)
	})
}

type DBUnit struct {
	tests []*Testing
}

func (d *DBUnit) NewDatabase(schema string, fixtures ...string) *gosql.DB {
	test := NewTest(schema)
	if len(fixtures) == 0 {
		fixtures = append(fixtures, filepath.Join(filepath.Dir(schema), "fixtures"))
	}
	test.Load(fixtures...)
	d.tests = append(d.tests, test)
	return test.DB()
}

func (d *DBUnit) drop() {
	for _, test := range d.tests {
		test.Drop()
	}
}

func New(t *testing.T, f func(d *DBUnit)) {
	dt := &DBUnit{}
	t.Cleanup(func() {
		dt.drop()
	})

	f(dt)
}
