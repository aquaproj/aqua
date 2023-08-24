package generate

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) readGeneratedPkgsFromFile(ctx context.Context, logE *logrus.Entry, param *config.Param, outputPkgs []*aqua.Package, m map[string]*FindingPackage) ([]*aqua.Package, error) {
	var file io.Reader
	if param.File == "-" {
		file = c.stdin
	} else {
		f, err := c.fs.Open(param.File)
		if err != nil {
			return nil, fmt.Errorf("open the package list file: %w", err)
		}
		defer f.Close()
		file = f
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		txt := getGeneratePkg(scanner.Text())
		key, version, _ := strings.Cut(txt, "@")
		findingPkg, ok := m[key]
		if !ok {
			return nil, logerr.WithFields(errUnknownPkg, logrus.Fields{"package_name": txt}) //nolint:wrapcheck
		}
		findingPkg.Version = version
		outputPkg := c.getOutputtedPkg(ctx, logE, param, findingPkg)
		outputPkgs = append(outputPkgs, outputPkg)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read the file: %w", err)
	}
	return outputPkgs, nil
}
