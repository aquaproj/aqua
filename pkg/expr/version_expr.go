package expr

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type Reader struct {
	pwd string
	fs  afero.Fs
}

func EvalVersionExpr(fs afero.Fs, pwd string, expression string) (string, error) {
	r := Reader{fs: fs, pwd: pwd}
	compiled, err := expr.Compile(expression, expr.Env(map[string]any{
		"readFile": r.readFile,
		"readJSON": r.readJSON,
		"readYAML": r.readYAML,
	}))
	if err != nil {
		return "", fmt.Errorf("parse the expression: %w", err)
	}
	a, err := expr.Run(compiled, map[string]any{
		"readFile": r.readFile,
		"readJSON": r.readJSON,
		"readYAML": r.readYAML,
	})
	if err != nil {
		return "", fmt.Errorf("evaluate the expression: %w", err)
	}
	s, ok := a.(string)
	if !ok {
		return "", errMustBeBoolean
	}
	return s, nil
}

func (r *Reader) read(s string) []byte {
	if !filepath.IsAbs(s) {
		s = filepath.Join(r.pwd, s)
	}
	b, err := afero.ReadFile(r.fs, s)
	if err != nil {
		panic(err)
	}
	return b
}

func (r *Reader) readFile(s string) string {
	return strings.TrimSpace(string(r.read(s)))
}

func (r *Reader) readJSON(s string) any {
	b := r.read(s)
	var a any
	if err := json.Unmarshal(b, &a); err != nil {
		panic(err)
	}
	return a
}

func (r *Reader) readYAML(s string) any {
	b := r.read(s)
	var a any
	if err := yaml.Unmarshal(b, &a); err != nil {
		panic(err)
	}
	return a
}
