package checksum

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
)

var (
	errUnknownChecksumFileFormat = errors.New("checksum file format is unknown")
	errNoChecksumExtracted       = errors.New("no checksum is extracted")
)

type FileParser struct{}

func (parser *FileParser) ParseChecksumFile(content string, pkg *config.Package) (map[string]string, string, error) {
	m, s, err := parser.parseChecksumFile(content, pkg)
	if err != nil {
		return nil, "", err
	}
	if len(m) == 0 && s == "" {
		return nil, "", errNoChecksumExtracted
	}
	return m, s, nil
}

func (parser *FileParser) parseChecksumFile(content string, pkg *config.Package) (map[string]string, string, error) {
	switch pkg.PackageInfo.Checksum.FileFormat {
	case "regexp":
		return parser.parseRegex(content, pkg)
	case "":
		return parser.parseDefault(content)
	}
	return nil, "", errUnknownChecksumFileFormat
}

func (parser *FileParser) parseDefault(content string) (map[string]string, string, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 1 && !strings.Contains(lines[0], " ") {
		return nil, lines[0], nil
	}
	m := make(map[string]string, len(lines))
	for _, line := range lines {
		idx := strings.Index(line, " ")
		if idx == -1 {
			continue
		}
		m[path.Base(strings.TrimSpace(line[idx:]))] = line[:idx]
	}
	if len(m) == 0 {
		return nil, "", errNoChecksumExtracted
	}
	return m, "", nil
}

func (parser *FileParser) parseRegex(content string, pkg *config.Package) (map[string]string, string, error) {
	checksumRegexp, err := regexp.Compile(pkg.PackageInfo.Checksum.Pattern.Checksum)
	if err != nil {
		return nil, "", fmt.Errorf("compile the checksum regular expression: %w", err)
	}

	if pkg.PackageInfo.Checksum.Pattern.File == "" {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			chksum := parser.extractByRegex(line, checksumRegexp)
			if chksum == "" {
				continue
			}
			return nil, chksum, nil
		}
	}

	fileRegexp, err := regexp.Compile(pkg.PackageInfo.Checksum.Pattern.File)
	if err != nil {
		return nil, "", fmt.Errorf("compile the checksum file name regular expression: %w", err)
	}
	lines := strings.Split(content, "\n")
	m := make(map[string]string, len(lines))
	for _, line := range lines {
		chksum := parser.extractByRegex(line, checksumRegexp)
		if chksum == "" {
			continue
		}
		file := parser.extractByRegex(line, fileRegexp)
		if file == "" {
			continue
		}
		m[file] = chksum
	}
	return m, "", nil
}

func (parser *FileParser) extractByRegex(line string, pattern *regexp.Regexp) string {
	if match := pattern.FindStringSubmatch(line); match != nil {
		if len(match) > 1 {
			return match[1]
		}
	}
	return ""
}
