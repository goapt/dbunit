package dbunit

import "fmt"

var Debug = false

type log struct{}

func (l *log) Print(s string) {
	fmt.Printf("🐳 %s\n", s)
}

func (l *log) Debug(s string) {
	if Debug {
		fmt.Printf("[DEBUG] %s\n", s)
	}
}

var defaultLog = &log{}
