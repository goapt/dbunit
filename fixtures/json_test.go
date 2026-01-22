package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_recursiveToJSON(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		data := make(map[any]any, 0)
		data["user_id"] = 1
		data["name"] = "test"
		ret := recursiveToJSON(data)

		assert.Equal(t, jsonMap(map[string]any{
			"user_id": 1,
			"name":    "test",
		}), ret)
	})

	t.Run("slice", func(t *testing.T) {
		data := make([]any, 0)
		data = append(data, map[any]any{
			"user_id": 1,
			"name":    "test",
		})
		ret := recursiveToJSON(data)

		exp := jsonMap(map[string]any{
			"user_id": 1,
			"name":    "test",
		})
		exp2 := jsonArray{exp}
		assert.Equal(t, exp2, ret)
	})
}
