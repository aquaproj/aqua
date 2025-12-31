package checksum

import (
	"errors"
	"fmt"
	"log/slog"
	"path"
	"regexp"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

// showFileContent logs checksum file content for debugging when checksum is not found.
// If content exceeds 10000 characters, it truncates the output.
func showFileContent(logger *slog.Logger, content string) {
	if len(content) > 10000 { //nolint:mnd
		content = content[:10000]
		logger.Error("Checksum isn't found in a checksum file. Checksum file content (10000):\n" + content)
		return
	}
	logger.Error("Checksum isn't found in a checksum file. Checksum file content:\n" + content)
}

// GetChecksum extracts a checksum for the specified asset from checksum file content.
// It parses the checksum file according to the provided configuration and returns
// the checksum value for the given asset name.
func GetChecksum(logger *slog.Logger, assetName, checksumFileContent string, checksumConfig *registry.Checksum) (string, error) {
	m, s, err := ParseChecksumFile(checksumFileContent, checksumConfig)
	logger = logger.With("checksum_file_format", checksumConfig.FileFormat)
	if checksumConfig.Pattern != nil {
		logger = logger.With(
			"checksum_file_format", checksumConfig.FileFormat,
			"checksum_pattern_checksum", checksumConfig.Pattern.Checksum,
			"checksum_pattern_file", checksumConfig.Pattern.File,
		)
	}
	if err != nil {
		if errors.Is(err, ErrNoChecksumExtracted) {
			showFileContent(logger, checksumFileContent)
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
	showFileContent(logger, checksumFileContent)
	return "", ErrNoChecksumIsFound
}

// ParseChecksumFile parses checksum file content according to the provided configuration.
// It returns a map of asset names to checksums, a single checksum string (for raw format),
// and any error encountered during parsing.
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

// parseChecksumFile is the internal implementation for parsing checksum files.
// It handles different file formats: raw, regexp, and default.
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

// parseDefault parses checksum files in the default format where each line contains
// a checksum followed by a filename, separated by space or tab.
// If the content is a single line without separators, it's treated as a raw checksum.
func parseDefault(content string) (map[string]string, string, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 1 && !strings.Contains(lines[0], " ") && !strings.Contains(lines[0], "\t") {
		return nil, lines[0], nil
	}
	m := make(map[string]string, len(lines))
	for _, line := range lines {
		idx := strings.Index(line, " ")
		if idx == -1 {
			idx = strings.Index(line, "\t")
			if idx == -1 {
				continue
			}
		}
		m[strings.TrimPrefix(path.Base(strings.TrimSpace(line[idx:])), "*")] = line[:idx]
	}
	if len(m) == 0 {
		return nil, "", ErrNoChecksumExtracted
	}
	return m, "", nil
}

// parseRegex parses checksum files using regular expressions to extract checksums and filenames.
// If no file pattern is provided, it returns the first checksum found.
// Otherwise, it returns a map of filenames to their corresponding checksums.
func parseRegex(content string, checksumPattern *registry.ChecksumPattern) (map[string]string, string, error) {
	checksumRegexp, err := regexp.Compile(checksumPattern.Checksum)
	if err != nil {
		return nil, "", fmt.Errorf("compile the checksum regular expression: %w", err)
	}

	if checksumPattern.File == "" {
		lines := strings.SplitSeq(content, "\n")
		for line := range lines {
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

// extractByRegex extracts the first capture group from a line using the provided regular expression.
// If no match is found or no capture groups exist, it returns an empty string.
func extractByRegex(line string, pattern *regexp.Regexp) string {
	if match := pattern.FindStringSubmatch(line); match != nil {
		if len(match) > 1 {
			return match[1]
		}
	}
	return ""
}
