package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_recursiveToJSON(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		data := make(map[interface{}]interface{}, 0)
		data["user_id"] = 1
		data["name"] = "test"
		ret := recursiveToJSON(data)

		assert.Equal(t, jsonMap(map[string]interface{}{
			"user_id": 1,
			"name":    "test",
		}), ret)
	})

	t.Run("slice", func(t *testing.T) {
		data := make([]interface{}, 0)
		data = append(data, map[interface{}]interface{}{
			"user_id": 1,
			"name":    "test",
		})
		ret := recursiveToJSON(data)

		exp := jsonMap(map[string]interface{}{
			"user_id": 1,
			"name":    "test",
		})
		exp2 := jsonArray{exp}
		assert.Equal(t, exp2, ret)
	})
}
