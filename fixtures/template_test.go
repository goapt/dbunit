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
