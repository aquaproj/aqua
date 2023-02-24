package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func Compile(s string) (*template.Template, error) {
	// delete some functions for security reason
	fncs := sprig.TxtFuncMap()
	delete(fncs, "env")
	delete(fncs, "expandenv")
	delete(fncs, "getHostByName")
	return template.New("_").Funcs(fncs).Funcs(template.FuncMap{ //nolint:wrapcheck
		"trimV": func(s string) string {
			return strings.TrimPrefix(s, "v")
		},
	}).Parse(s)
}

func Execute(s string, input interface{}) (string, error) {
	tpl, err := Compile(s)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	return ExecuteTemplate(tpl, input)
}

func ExecuteTemplate(tpl *template.Template, input interface{}) (string, error) {
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, input); err != nil {
		return "", fmt.Errorf("render a template: %w", err)
	}
	return buf.String(), nil
}
