package genrgst

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
)

const template = `---
# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/aqua-generate-registry.json
# aqua - Declarative CLI Version Manager
# https://aquaproj.github.io/
package: %%PACKAGE%%
version_filter: not (Version matches "-rc$")
asset: not (Asset matches "-cli")
`

func (c *Controller) initConfig(args ...string) error {
	if len(args) == 0 {
		return errors.New("package name is required")
	}
	if err := afero.WriteFile(c.fs, "aqua-generate-registry.yaml", []byte(strings.Replace(template, "%%PACKAGE%%", args[0], 1)), osfile.FilePermission); err != nil {
		return fmt.Errorf("write aqua-generate-registry.yaml: %w", err)
	}
	return nil
}
