package dbunit

import "fmt"

var Debug = false

type log struct{}

func (l *log) Print(s string) {
	fmt.Println(fmt.Sprintf("🐳 %s", s))
}

func (l *log) Debug(s string) {
	if Debug {
		fmt.Println(fmt.Sprintf("[DEBUG] %s", s))
	}
}

var defaultLog = &log{}
