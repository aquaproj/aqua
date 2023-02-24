package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type Outputter struct {
	stdout io.Writer
	fs     afero.Fs
}

func New(stdout io.Writer, fs afero.Fs) *Outputter {
	return &Outputter{
		stdout: stdout,
		fs:     fs,
	}
}

type Param struct {
	Insert         bool
	Dest           string
	List           []*aqua.Package
	ConfigFilePath string
}

func (out *Outputter) Output(param *Param) error {
	if !param.Insert && param.Dest == "" {
		if err := yaml.NewEncoder(out.stdout).Encode(param.List); err != nil {
			return fmt.Errorf("output generated package configuration: %w", err)
		}
		return nil
	}

	if param.Dest == "" {
		return out.generateInsert(param.ConfigFilePath, param.List)
	}

	if _, err := out.fs.Stat(param.Dest); err == nil {
		return out.generateInsert(param.Dest, param.List)
	}

	f, err := out.fs.Create(param.Dest)
	if err != nil {
		return fmt.Errorf("create a file: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString("packages:\n  "); err != nil {
		return fmt.Errorf("write a string to a file %s: %w", param.Dest, err)
	}

	b, err := yaml.Marshal(param.List)
	if err != nil {
		return fmt.Errorf("marshal packages as YAML: %w", err)
	}
	if _, err := f.WriteString(strings.Join(strings.Split(strings.TrimSpace(string(b)), "\n"), "\n  ") + "\n"); err != nil {
		return fmt.Errorf("write a string to a file %s: %w", param.Dest, err)
	}
	return nil
}
