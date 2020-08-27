package dbunit

import (
	"io/ioutil"
	"os"
	"reflect"

	"gopkg.in/yaml.v2"

	"git.verystar.cn/gopkg/dbunit/fixtures"
)

func PluckWithFixture(filePath string, key string) []interface{} {
	var data = make([]map[string]interface{}, 0)
	if !isExists(filePath) {
		panic("file not exists:" + filePath)
	}
	d, _ := ioutil.ReadFile(filePath)
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
