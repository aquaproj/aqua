package expr

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/spf13/afero"
	"go.yaml.in/yaml/v3"
)

type Reader struct {
	pwd string
	fs  afero.Fs
}

const safeVersionPattern = `^v?\d+\.\d+(\.\d+)*[.-]?((alpha|beta|dev|rc)[.-]?)?\d*`

var safeVersionRegexp = regexp.MustCompile(safeVersionPattern)

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
		// Don't output error to prevent leaking sensitive information
		// Maybe malicious users tries to read a secret file
		return "", errors.New("evaluate the expression")
	}
	s, ok := a.(string)
	if !ok {
		return "", errMustBeString
	}
	// Restrict the value of version_expr to a semver for security reason.
	// This prevents secrets from being exposed.
	if !safeVersionRegexp.MatchString(s) {
		// Don't output the valuof of version_expr to prevent leaking sensitive information
		// Maybe malicious users tries to read a secret file
		return "", errors.New("the evaluation result of version_expr must match with " + safeVersionPattern)
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
		// Don't output error to prevent leaking sensitive information
		// Maybe malicious users tries to read a secret file
		panic("failed to unmarshal JSON")
	}
	return a
}

func (r *Reader) readYAML(s string) any {
	b := r.read(s)
	var a any
	if err := yaml.Unmarshal(b, &a); err != nil {
		// Don't output error to prevent leaking sensitive information
		// Maybe malicious users tries to read a secret file
		panic("failed to unmarshal YAML")
	}
	return a
}
