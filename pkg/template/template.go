package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/aquaproj/aqua/v2/pkg/asset"
)

func Compile(s string) (*template.Template, error) {
	// delete some functions for security reason
	fncs := sprig.TxtFuncMap()
	delete(fncs, "env")
	delete(fncs, "expandenv")
	delete(fncs, "getHostByName")
	return template.New("_").Funcs(fncs).Funcs(template.FuncMap{ //nolint:wrapcheck
		"trimAssetExt": func(s string) string {
			s, _ = asset.RemoveExtFromAsset(s)
			return s
		},
		"trimV": func(s string) string {
			return strings.TrimPrefix(s, "v")
		},
	}).Parse(s)
}

func Execute(s string, input any) (string, error) {
	tpl, err := Compile(s)
	if err != nil {
		return "", fmt.Errorf("parse a template: %w", err)
	}
	return ExecuteTemplate(tpl, input)
}

func ExecuteTemplate(tpl *template.Template, input any) (string, error) {
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, input); err != nil {
		return "", fmt.Errorf("render a template: %w", err)
	}
	return buf.String(), nil
}
