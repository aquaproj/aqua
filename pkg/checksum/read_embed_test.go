package checksum_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
)

func TestReadEmbeddedTool(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name         string
		aquaBytes    []byte
		checksumInfo []byte
		expectPanic  bool
	}{
		{
			name: "valid embedded tool",
			aquaBytes: []byte(`
packages:
  - name: test-tool
    version: v1.0.0
`),
			checksumInfo: []byte(`{
				"checksums": [
					{
						"id": "test-id",
						"checksum": "abc123",
						"algorithm": "sha256"
					}
				]
			}`),
			expectPanic: false,
		},
		{
			name: "invalid aqua YAML",
			aquaBytes: []byte(`
invalid yaml syntax: [
  unclosed bracket
`),
			checksumInfo: []byte(`{
				"checksums": []
			}`),
			expectPanic: true,
		},
		{
			name: "invalid checksum JSON",
			aquaBytes: []byte(`
packages:
  - name: test-tool
    version: v1.0.0
`),
			checksumInfo: []byte(`{
				"checksums": [
					"invalid": json
				]
			}`),
			expectPanic: true,
		},
		{
			name: "empty aqua config",
			aquaBytes: []byte(`
packages: []
`),
			checksumInfo: []byte(`{"checksums": []}`),
			expectPanic:  true, // Will panic when accessing cfg.Packages[0]
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()

			checksums := checksum.New()
			var version string
			var panicked bool

			defer func() {
				if r := recover(); r != nil {
					panicked = true
					if !d.expectPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			version = checksum.ReadEmbeddedTool(checksums, d.aquaBytes, d.checksumInfo)

			if d.expectPanic && !panicked {
				t.Error("Expected panic but function completed normally")
				return
			}

			if panicked {
				return // Test complete for panic case
			}

			// Basic validation that we got valid objects back
			if version == "" {
				t.Error("Expected non-empty version string")
			}
			if checksums == nil {
				t.Error("Expected non-nil checksums")
			}
		})
	}
}

func TestReadEmbeddedToolWithValidData(t *testing.T) {
	t.Parallel()

	aquaBytes := []byte(`
packages:
  - name: aqua
    version: v2.0.0
    registry: aquaproj/aqua-registry
`)

	checksumBytes := []byte(`{
		"checksums": [
			{
				"id": "github_release/github.com/aquaproj/aqua/v2.0.0/aqua_linux_amd64.tar.gz",
				"checksum": "abcdef1234567890",
				"algorithm": "sha256"
			},
			{
				"id": "github_release/github.com/aquaproj/aqua/v2.0.0/aqua_darwin_amd64.tar.gz",
				"checksum": "1234567890abcdef",
				"algorithm": "sha256"
			}
		]
	}`)

	checksums := checksum.New()
	version := checksum.ReadEmbeddedTool(checksums, aquaBytes, checksumBytes)

	// Verify version
	if version != "v2.0.0" {
		t.Errorf("Expected version 'v2.0.0', got '%s'", version)
	}

	// Verify checksums - check that we can retrieve the checksums we set
	chk1 := checksums.Get("github_release/github.com/aquaproj/aqua/v2.0.0/aqua_linux_amd64.tar.gz")
	if chk1 == nil {
		t.Error("Expected to find Linux checksum")
	} else {
		expectedChecksum := "abcdef1234567890" // Actual case from Set method
		if chk1.Checksum != expectedChecksum {
			t.Errorf("Expected checksum '%s', got '%s'", expectedChecksum, chk1.Checksum)
		}
		if chk1.Algorithm != "sha256" {
			t.Errorf("Expected algorithm 'sha256', got '%s'", chk1.Algorithm)
		}
	}

	chk2 := checksums.Get("github_release/github.com/aquaproj/aqua/v2.0.0/aqua_darwin_amd64.tar.gz")
	if chk2 == nil {
		t.Error("Expected to find Darwin checksum")
	} else {
		expectedChecksum := "1234567890abcdef" // Actual case from Set method
		if chk2.Checksum != expectedChecksum {
			t.Errorf("Expected checksum '%s', got '%s'", expectedChecksum, chk2.Checksum)
		}
	}
}
