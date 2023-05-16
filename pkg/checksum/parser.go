package checksum

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

func showFileContent(logE *logrus.Entry, content string) {
	if len(content) > 10000 { //nolint:gomnd
		content = content[:10000]
		logE.Error(fmt.Sprintf("Checksum isn't found in a checksum file. Checksum file content (10000):\n%s", content))
		return
	}
	logE.Error(fmt.Sprintf("Checksum isn't found in a checksum file. Checksum file content:\n%s", content))
}

func GetChecksum(logE *logrus.Entry, assetName, checksumFileContent string, checksumConfig *registry.Checksum) (string, error) {
	m, s, err := ParseChecksumFile(checksumFileContent, checksumConfig)
	logE = logE.WithField("checksum_file_format", checksumConfig.FileFormat)
	if checksumConfig.Pattern != nil {
		logE = logE.WithFields(logrus.Fields{
			"checksum_pattern_checksum": checksumConfig.Pattern.Checksum,
			"checksum_pattern_file":     checksumConfig.Pattern.File,
		})
	}
	if err != nil {
		if errors.Is(err, ErrNoChecksumExtracted) {
			showFileContent(logE, checksumFileContent)
		}
		return "", fmt.Errorf("parse a checksum file: %w", err)
	}
	if s != "" {
		return s, nil
	}
	a, ok := m[assetName]
	if ok {
		return a, nil
	}
	showFileContent(logE, checksumFileContent)
	return "", ErrNoChecksumIsFound
}

func ParseChecksumFile(content string, checksumConfig *registry.Checksum) (map[string]string, string, error) {
	m, s, err := parseChecksumFile(content, checksumConfig)
	if err != nil {
		return nil, "", err
	}
	if len(m) == 0 && s == "" {
		return nil, "", ErrNoChecksumExtracted
	}
	return m, s, nil
}

func parseChecksumFile(content string, checksumConfig *registry.Checksum) (map[string]string, string, error) {
	switch checksumConfig.FileFormat {
	case "raw":
		return nil, strings.TrimSpace(content), nil
	case "regexp":
		return parseRegex(content, checksumConfig.Pattern)
	case "":
		return parseDefault(content)
	}
	return nil, "", errUnknownChecksumFileFormat
}

func parseDefault(content string) (map[string]string, string, error) {
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
		m[strings.TrimPrefix(path.Base(strings.TrimSpace(line[idx:])), "*")] = line[:idx]
	}
	if len(m) == 0 {
		return nil, "", ErrNoChecksumExtracted
	}
	return m, "", nil
}

func parseRegex(content string, checksumPattern *registry.ChecksumPattern) (map[string]string, string, error) {
	checksumRegexp, err := regexp.Compile(checksumPattern.Checksum)
	if err != nil {
		return nil, "", fmt.Errorf("compile the checksum regular expression: %w", err)
	}

	if checksumPattern.File == "" {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			chksum := extractByRegex(line, checksumRegexp)
			if chksum == "" {
				continue
			}
			return nil, chksum, nil
		}
	}

	fileRegexp, err := regexp.Compile(checksumPattern.File)
	if err != nil {
		return nil, "", fmt.Errorf("compile the checksum file name regular expression: %w", err)
	}
	lines := strings.Split(content, "\n")
	m := make(map[string]string, len(lines))
	for _, line := range lines {
		chksum := extractByRegex(line, checksumRegexp)
		if chksum == "" {
			continue
		}
		file := extractByRegex(line, fileRegexp)
		if file == "" {
			continue
		}
		m[file] = chksum
	}
	return m, "", nil
}

func extractByRegex(line string, pattern *regexp.Regexp) string {
	if match := pattern.FindStringSubmatch(line); match != nil {
		if len(match) > 1 {
			return match[1]
		}
	}
	return ""
}
