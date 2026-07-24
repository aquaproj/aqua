package output

import (
	"fmt"
	"io"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	goccyYAML "github.com/goccy/go-yaml"
	"go.yaml.in/yaml/v2"
)

type Outputter struct {
	stdout io.Writer
}

func New(stdout io.Writer) *Outputter {
	return &Outputter{
		stdout: stdout,
	}
}

type Param struct {
	Insert         bool
	Dest           string
	List           []*aqua.Package
	ConfigFilePath string
}

func (o *Outputter) Output(param *Param) error {
	if !param.Insert && param.Dest == "" {
		if err := yaml.NewEncoder(o.stdout).Encode(param.List); err != nil {
			return fmt.Errorf("output generated package configuration: %w", err)
		}
		return nil
	}

	if param.Dest == "" {
		return o.generateInsert(param.ConfigFilePath, param.List)
	}

	if _, err := os.Stat(param.Dest); err == nil {
		return o.generateInsert(param.Dest, param.List)
	}

	f, err := os.Create(param.Dest)
	if err != nil {
		return fmt.Errorf("create a file: %w", err)
	}
	defer f.Close()

	if err := goccyYAML.NewEncoder(f, goccyYAML.IndentSequence(true)).Encode(map[string]any{
		"packages": param.List,
	}); err != nil {
		return fmt.Errorf("encode YAML: %w", err)
	}
	return nil
}
