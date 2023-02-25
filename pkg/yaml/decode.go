package yaml

import (
	"fmt"

	"github.com/spf13/afero"
	yamlV2 "gopkg.in/yaml.v2"
	"gopkg.in/yaml.v3"
)

type Decoder struct {
	fs afero.Fs
}

func NewDecoder(fs afero.Fs) *Decoder {
	return &Decoder{
		fs: fs,
	}
}

// gopkg.in/yaml.v3 can't parse YAML which has duplicated keys, but aqua should allow invalid YAML as much as possible.
// So aqua uses gopkg.in/yaml.v3 mainly and uses gopkg.in/yaml.v2 only when gopkg.in/yaml.v3 can't parse YAML.

func (decoder *Decoder) ReadFile(p string, dest interface{}) error {
	fileV3, err := decoder.fs.Open(p)
	if err != nil {
		return fmt.Errorf("open a YAML file: %w", err)
	}
	defer fileV3.Close()
	errV3 := yaml.NewDecoder(fileV3).Decode(dest)
	if errV3 == nil {
		return nil
	}

	fileV2, err := decoder.fs.Open(p)
	if err != nil {
		return fmt.Errorf("open a YAML file: %w", err)
	}
	defer fileV2.Close()

	if err := yamlV2.NewDecoder(fileV2).Decode(dest); err == nil {
		return nil
	}
	return fmt.Errorf("parse a YAML file: %w", errV3)
}

func Unmarshal(b []byte, dest interface{}) error {
	errV3 := yaml.Unmarshal(b, dest)
	if errV3 == nil {
		return nil
	}
	if err := yamlV2.Unmarshal(b, dest); err == nil {
		return nil
	}
	return errV3 //nolint:wrapcheck
}
