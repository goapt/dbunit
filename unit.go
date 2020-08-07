package dbunit

import (
	"os"
	"reflect"
)

func Pluck(data []map[string]interface{}, key string) []interface{} {
	s := make([]interface{}, len(data))
	for k, v := range data {
		s[k] = v[key]
	}
	return unique(s)
}

func unique(s []interface{}) []interface{} {
	ns := make([]interface{}, 0)
	for _, v := range s {
		isDuplicate := false
		for _, nv := range ns {
			if reflect.DeepEqual(nv, v) {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			ns = append(ns, v)
		}
	}
	return ns
}

func isExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
