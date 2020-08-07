package fixtures

import (
	"testing"
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
