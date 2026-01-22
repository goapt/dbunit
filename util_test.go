package dbunit

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluckWithFixture(t *testing.T) {
	pd := PluckWithFixture("./testdata/fixtures/users.yml", "id")
	assert.Equal(t, []any{1, 2}, pd)
}

func TestPluck(t *testing.T) {
	data := make([]map[string]any, 0)
	data = append(data, map[string]any{
		"user_id": 1,
	})
	data = append(data, map[string]any{
		"user_id": 2,
	})

	pd := Pluck(data, "user_id")
	assert.Equal(t, []any{1, 2}, pd)
}

func Test_unique(t *testing.T) {
	s := []any{1, 2, 2, 3, 2, 6, 3, 7}
	s2 := []any{1, 2, 3, 6, 7}
	if !reflect.DeepEqual(unique(s), s2) {
		t.Error("unique error")
	}
}
