package fixtures

import (
	"bytes"
	"text/template"
	"time"
)

type Template struct {
	templateFuncs      template.FuncMap
	templateLeftDelim  string
	templateRightDelim string
	templateOptions    []string
	templateData       interface{}
}

func NewTemplate() *Template {
	l := &Template{
		templateLeftDelim:  "{{",
		templateRightDelim: "}}",
		templateOptions:    []string{"missingkey=zero"},
	}

	l.templateFuncs = make(template.FuncMap)
	l.templateFuncs["now"] = func() string {
		return time.Now().Format(time.RFC3339)
	}
	return l
}

func (t *Template) Parse(content []byte) ([]byte, error) {
	tpl := template.New("").
		Funcs(t.templateFuncs).
		Delims(t.templateLeftDelim, t.templateRightDelim).
		Option(t.templateOptions...)
	tpl, err := tpl.Parse(string(content))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, t.templateData); err != nil {
		return nil, err
	}

	content = buffer.Bytes()
	return content, nil
}
