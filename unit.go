package dbunit

import (
	"os"
	"reflect"

	"gopkg.in/yaml.v2"

	"github.com/goapt/dbunit/fixtures"
)

func PluckWithFixture(filePath string, key string) []any {
	var data = make([]map[string]any, 0)
	if !isExists(filePath) {
		panic("file not exists:" + filePath)
	}
	d, _ := os.ReadFile(filePath)
	tpl := fixtures.NewTemplate()
	d, err := tpl.Parse(d)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(d, &data)
	if err != nil {
		panic(err)
	}

	return Pluck(data, key)
}

func Pluck(data []map[string]any, key string) []any {
	s := make([]any, len(data))
	for k, v := range data {
		s[k] = v[key]
	}
	return unique(s)
}

func unique(s []any) []any {
	ns := make([]any, 0)
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
