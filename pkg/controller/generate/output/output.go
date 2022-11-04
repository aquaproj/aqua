package output

import (
	"fmt"
	"io"

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

	if param.Dest != "" {
		if _, err := out.fs.Stat(param.Dest); err != nil {
			if err := afero.WriteFile(out.fs, param.Dest, []byte("packages:\n\n"), 0o644); err != nil { //nolint:gomnd
				return fmt.Errorf("create a file: %w", err)
			}
		}
		return out.generateInsert(param.Dest, param.List)
	}

	return out.generateInsert(param.ConfigFilePath, param.List)
}
