package generate

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (c *Controller) readGeneratedPkgsFromFile(ctx context.Context, logger *slog.Logger, param *config.Param, outputPkgs []*config.Package, m map[string]*fuzzyfinder.Package) ([]*config.Package, error) {
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
			return nil, slogerr.With(errUnknownPkg, "package_name", txt) //nolint:wrapcheck
		}
		findingPkg.Version = version
		outputPkg := c.getOutputtedPkg(ctx, logger, param, findingPkg)
		outputPkgs = append(outputPkgs, outputPkg)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read the file: %w", err)
	}
	return outputPkgs, nil
}
