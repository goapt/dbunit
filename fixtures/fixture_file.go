package fixtures

import (
	"path/filepath"
	"strings"
)

type fixtureFile struct {
	path      string
	fileName  string
	content   []byte
	insertSQL insertSQL
}

func (f *fixtureFile) fileNameWithoutExtension() string {
	return strings.Replace(f.fileName, filepath.Ext(f.fileName), "", 1)
}
