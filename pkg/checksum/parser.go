package checksum

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
)

var errUnknownChecksumFileFormat = errors.New("checksum file format is unknown")

type FileParser struct{}

func (parser *FileParser) ParseChecksumFile(content string, pkg *config.Package) (map[string]string, error) {
	switch pkg.PackageInfo.Checksum.FileFormat { //nolint:gocritic
	case "regexp":
		return parser.parseRegex(content, pkg)
	}
	return nil, errUnknownChecksumFileFormat
}

func (parser *FileParser) parseRegex(content string, pkg *config.Package) (map[string]string, error) {
	checksumRegexp, err := regexp.Compile(pkg.PackageInfo.Checksum.Pattern.Checksum)
	if err != nil {
		return nil, fmt.Errorf("compile the checksum regular expression: %w", err)
	}

	fileRegexp, err := regexp.Compile(pkg.PackageInfo.Checksum.Pattern.File)
	if err != nil {
		return nil, fmt.Errorf("compile the checksum file name regular expression: %w", err)
	}
	pattern := &RegexPattern{
		File:     fileRegexp,
		Checksum: checksumRegexp,
	}
	lines := strings.Split(content, "\n")
	m := make(map[string]string, len(lines))
	for _, line := range lines {
		chksum, file := parser.extractRegex(line, pattern)
		if file == "" || chksum == "" {
			continue
		}
		m[file] = chksum
	}
	return m, nil
}

type RegexPattern struct {
	Checksum *regexp.Regexp
	File     *regexp.Regexp
}

func (parser *FileParser) extractRegex(line string, pattern *RegexPattern) (string, string) {
	checksum := ""
	if match := pattern.Checksum.FindStringSubmatch(line); match != nil {
		if len(match) > 1 {
			checksum = match[1]
		}
	}

	file := ""
	if match := pattern.File.FindStringSubmatch(line); match != nil {
		if len(match) > 1 {
			file = match[1]
		}
	}

	return checksum, file
}
