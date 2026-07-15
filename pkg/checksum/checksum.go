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
	"math"
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
	case algoMD5:
		return md5.New(), nil //nolint:gosec
	case algoSHA256:
		return sha256.New(), nil
	case algoSHA512:
		return sha512.New(), nil
	case algoSHA1:
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

// isSignatureFile reports whether a filename looks like a signature file.
func isSignatureFile(filename string) bool {
	s := strings.ToLower(filename)
	for _, suffix := range []string{"sig", "asc", "pem", "cert", "bundle", "sigstore", "sigstore.json", "minisig", "gpg", "gpgsig"} {
		if strings.HasSuffix(s, "."+suffix) {
			return true
		}
	}
	return false
}

// matchAlgorithm returns the checksum algorithm corresponding to a filename and its
// priority, or an empty string and a negative priority if the filename does not look
// like a checksum file. The priority is the index in the inlined patterns below
// (lower index = higher preference: sha512 > sha256 > sha1 > md5).
func matchAlgorithm(filename string) (string, int) {
	if isSignatureFile(filename) {
		return "", math.MaxInt
	}
	s := strings.ToLower(filename)
	// Ordered by preference: sha512 > sha256 > sha1 > md5 (lower index = higher priority).
	for i, p := range []struct {
		words     []string
		algorithm string
	}{
		{words: []string{algoSHA512, "shasums512"}, algorithm: algoSHA512},
		{words: []string{algoSHA256, "shasums", "checksum"}, algorithm: algoSHA256},
		{words: []string{algoSHA1, "shasum1"}, algorithm: algoSHA1},
		{words: []string{algoMD5}, algorithm: algoMD5},
	} {
		for _, w := range p.words {
			if strings.Contains(s, w) {
				return p.algorithm, i
			}
		}
	}
	return "", math.MaxInt
}

// IsChecksumFile reports whether a single filename appears to be a checksum file.
// Signature files are never treated as checksum files.
func IsChecksumFile(filename string) bool {
	algorithm, _ := matchAlgorithm(filename)
	return algorithm != ""
}

// GetChecksumConfigFromFilename analyzes a list of asset names to determine the best
// checksum configuration. It identifies checksum files by their filename patterns,
// excludes signature files, and selects the one with the most preferred hash algorithm
// (sha512 > sha256 > sha1 > md5).
// It returns the checksum configuration for github_release type, or nil if no asset matches.
// The returned checksum's Asset field uses convertChecksumFileName.
func GetChecksumConfigFromFilename(assets []string, version string) *registry.Checksum {
	_, chksum := GetChecksumConfigFromFilenameWithName(assets, version)
	return chksum
}

// GetChecksumConfigFromFilenameWithName is like GetChecksumConfigFromFilename but also
// returns the original matched asset name alongside the checksum configuration.
func GetChecksumConfigFromFilenameWithName(assets []string, version string) (string, *registry.Checksum) {
	type candidate struct {
		filename  string
		algorithm string
		priority  int
	}
	var best *candidate

	for _, filename := range assets {
		algorithm, prio := matchAlgorithm(filename)
		if algorithm == "" {
			continue
		}
		if best == nil || prio < best.priority {
			best = &candidate{
				filename:  filename,
				algorithm: algorithm,
				priority:  prio,
			}
		}
	}

	if best == nil {
		return "", nil
	}

	return best.filename, &registry.Checksum{
		Type:      "github_release",
		Algorithm: best.algorithm,
		Asset:     convertChecksumFileName(best.filename, version),
	}
}
