package checksum

import (
	"crypto/md5"  //nolint:gosec
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/spf13/afero"
)

// Calculator provides checksum calculation functionality for files and streams.
// It supports multiple hash algorithms including MD5, SHA1, SHA256, and SHA512.
type Calculator struct{}

// NewCalculator creates a new checksum calculator instance.
// Returns a Calculator that can compute hashes for files and streams.
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Calculate computes the checksum of a file using the specified algorithm.
// It opens the file from the filesystem and delegates to CalculateReader.
// Returns the hexadecimal representation of the computed hash.
func (*Calculator) Calculate(fs afero.Fs, filename, algorithm string) (string, error) {
	f, err := fs.Open(filename)
	if err != nil {
		return "", fmt.Errorf("open a file to calculate the checksum: %w", err)
	}
	defer f.Close()
	return CalculateReader(f, algorithm)
}

// CalculateReader computes the checksum of data from an io.Reader using the specified algorithm.
// It reads all data from the reader and computes the hash using the appropriate algorithm.
// Returns the hexadecimal representation of the computed hash.
func CalculateReader(file io.Reader, algorithm string) (string, error) {
	h, err := getHash(algorithm)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("copy an io.Reader to hash object: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// getHash creates a hash.Hash instance for the specified algorithm.
// Supports md5, sha1, sha256, and sha512 algorithms.
// Returns an error for empty or unsupported algorithm names.
func getHash(algorithm string) (hash.Hash, error) {
	switch algorithm {
	case "md5":
		return md5.New(), nil //nolint:gosec
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	case "sha1":
		return sha1.New(), nil //nolint:gosec
	case "":
		return nil, errors.New("algorithm is required")
	default:
		return nil, errors.New("unsupported algorithm")
	}
}

// convertChecksumFileName converts a checksum filename to a template format.
// It replaces version strings with template placeholders for dynamic generation.
// Handles both prefixed (v1.0.0) and non-prefixed (1.0.0) version formats.
func convertChecksumFileName(filename, version string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(filename, version, "{{.Version}}"),
		strings.TrimPrefix(version, "v"), "{{trimV .Version}}")
}

// GetChecksumConfigFromFilename analyzes a filename to determine checksum configuration.
// It identifies the hash algorithm based on filename patterns and excludes signature files.
// Returns a checksum configuration for github_release type or nil if no pattern matches.
func GetChecksumConfigFromFilename(filename, version string) *registry.Checksum {
	s := strings.ToLower(filename)
	for _, suffix := range []string{"sig", "asc", "pem", "bundle"} {
		if strings.HasSuffix(s, "."+suffix) {
			return nil
		}
	}
	arr := []struct {
		words     []string
		algorithm string
	}{
		{
			words:     []string{"sha512", "shasums512"},
			algorithm: "sha512",
		},
		{
			words:     []string{"md5"},
			algorithm: "md5",
		},
		{
			words:     []string{"sha1", "shasum1"},
			algorithm: "sha1",
		},
		{
			words:     []string{"sha256", "shasums", "checksum"},
			algorithm: "sha256",
		},
	}
	for _, a := range arr {
		for _, w := range a.words {
			if strings.Contains(s, w) {
				return &registry.Checksum{
					Type:      "github_release",
					Algorithm: a.algorithm,
					Asset:     convertChecksumFileName(filename, version),
				}
			}
		}
	}
	return nil
}
