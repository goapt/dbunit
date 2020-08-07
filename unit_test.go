package dbunit

import (
	"reflect"
	"testing"
)

func Test_unique(t *testing.T) {
	s := []interface{}{1, 2, 2, 3, 2, 6, 3, 7}
	s2 := []interface{}{1, 2, 3, 6, 7}
	if !reflect.DeepEqual(unique(s), s2) {
		t.Error("unique error")
	}
}
