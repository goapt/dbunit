package fixtures

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTemplate_processFileTemplate(t *testing.T) {
	tpl := NewTemplate()
	content := []byte(`created_at: {{now}}`)
	content, err := tpl.Parse(content)
	assert.NoError(t, err)
	assert.Contains(t, string(content), time.Now().Format("2006-01-02"))
}

// approximate
func timeEq(t *testing.T, t1 time.Time, t2 time.Time) {
	if t1.Before(t2) {
		assert.True(t, t2.Sub(t1) < time.Second*2)
	} else {
		assert.True(t, t1.Sub(t2) < time.Second*2)
	}
}

func TestTemplate_processFileTemplate2(t *testing.T) {
	tpl := NewTemplate()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	test := []struct {
		expr   string
		expect time.Time
	}{
		{
			expr:   `{{now}}`,
			expect: time.Now(),
		},
		{
			expr:   `{{now 1 "day"}}`,
			expect: time.Now().Add(time.Hour * 24),
		},
		{
			expr:   `{{now -2 "day"}}`,
			expect: time.Now().Add(time.Hour * 24 * -2),
		},
		{
			expr:   `{{now 1 "minute"}}`,
			expect: time.Now().Add(time.Minute),
		},
		{
			expr:   `{{now 100 "second"}}`,
			expect: time.Now().Add(time.Second * 100),
		},
	}

	for _, tt := range test {
		content := []byte(tt.expr)
		content, err := tpl.Parse(content)
		assert.NoError(t, err)
		dt, err := tryStrToDate(loc, string(content))
		assert.NoError(t, err)
		timeEq(t, tt.expect, dt)
	}
}
