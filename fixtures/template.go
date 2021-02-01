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
	l.templateFuncs["now"] = func(args ...interface{}) string {
		if len(args) == 0 {
			return time.Now().Format(time.RFC3339)
		}
		num := time.Duration(args[0].(int))
		unit := args[1].(string)

		var duration time.Duration
		switch unit {
		case "day":
			duration = time.Hour * 24 * num
		case "minute":
			duration = time.Minute * num
		case "second":
			duration = time.Second * num
		default:
			duration = time.Second * num
		}
		return time.Now().Add(duration).Format(time.RFC3339)
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
