package controller

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type Template struct {
	raw      string
	template *template.Template
}

func NewTemplate(raw string) *Template {
	return &Template{
		raw: raw,
	}
}

func newT(s string) (*template.Template, error) {
	return template.New("_").Funcs(sprig.TxtFuncMap()).Funcs(template.FuncMap{ //nolint:wrapcheck
		"trimV": func(s string) string {
			return strings.TrimPrefix(s, "v")
		},
	}).Parse(s)
}

func (tpl *Template) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return err
	}
	tpl.raw = raw
	return nil
}

func (tpl *Template) Parse() error {
	if tpl.template != nil {
		return nil
	}
	a, err := newT(tpl.raw)
	if err != nil {
		return err
	}
	tpl.template = a
	return nil
}

func (tpl *Template) Execute(param interface{}) (string, error) {
	if err := tpl.Parse(); err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tpl.template.Execute(buf, param); err != nil {
		return "", fmt.Errorf("render a template: %w", err)
	}
	return buf.String(), nil
}
