package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fixtureFile_fileNameWithoutExtension(t *testing.T) {
	f := &fixtureFile{
		fileName: "orders.yml",
	}
	assert.Equal(t, "orders", f.fileNameWithoutExtension())
}
