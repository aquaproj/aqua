package checksum_test

import (
	"sort"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/google/go-cmp/cmp"
)

func TestGetRuntimesFromSupportedEnvs(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name              string
		pkgSupportedEnvs  []string
		confSupportedEnvs []string
		expected          []*runtime.Runtime
	}{
		{
			name:              "matching environments",
			pkgSupportedEnvs:  []string{"linux/amd64", "darwin/amd64", "windows/amd64"},
			confSupportedEnvs: []string{"linux", "darwin"},
			expected: []*runtime.Runtime{
				{GOOS: "linux", GOARCH: "amd64"},
				{GOOS: "darwin", GOARCH: "amd64"},
			},
		},
		{
			name:              "no matching environments",
			pkgSupportedEnvs:  []string{"linux/arm64", "darwin/arm64"},
			confSupportedEnvs: []string{"windows"},
			expected:          []*runtime.Runtime{},
		},
		{
			name:              "empty package environments",
			pkgSupportedEnvs:  nil,
			confSupportedEnvs: []string{"linux", "darwin"},
			expected: []*runtime.Runtime{
				{GOOS: "linux", GOARCH: "amd64"},
				{GOOS: "linux", GOARCH: "arm64"},
				{GOOS: "darwin", GOARCH: "amd64"},
				{GOOS: "darwin", GOARCH: "arm64"},
			},
		},
		{
			name:              "empty config environments returns all package envs",
			pkgSupportedEnvs:  []string{"linux/amd64", "darwin/amd64"},
			confSupportedEnvs: nil,
			expected: []*runtime.Runtime{
				{GOOS: "linux", GOARCH: "amd64"},
				{GOOS: "darwin", GOARCH: "amd64"},
			},
		},
		{
			name:              "partial matches",
			pkgSupportedEnvs:  []string{"linux/amd64", "linux/arm64", "windows/amd64"},
			confSupportedEnvs: []string{"linux"},
			expected: []*runtime.Runtime{
				{GOOS: "linux", GOARCH: "amd64"},
				{GOOS: "linux", GOARCH: "arm64"},
			},
		},
		{
			name:              "exact arch match",
			pkgSupportedEnvs:  []string{"linux/amd64", "darwin/arm64", "windows/amd64"},
			confSupportedEnvs: []string{"amd64"},
			expected: []*runtime.Runtime{
				{GOOS: "linux", GOARCH: "amd64"},
				{GOOS: "windows", GOARCH: "amd64"},
			},
		},
		{
			name:              "mixed OS and arch filters",
			pkgSupportedEnvs:  []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64"},
			confSupportedEnvs: []string{"linux", "arm64"},
			expected: []*runtime.Runtime{
				{GOOS: "linux", GOARCH: "amd64"},
				{GOOS: "linux", GOARCH: "arm64"},
				{GOOS: "darwin", GOARCH: "arm64"},
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result, err := checksum.GetRuntimesFromSupportedEnvs(
				d.confSupportedEnvs,
				d.pkgSupportedEnvs,
			)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Sort both slices for comparison since order is not guaranteed
			sort.Slice(result, func(i, j int) bool {
				if result[i].GOOS != result[j].GOOS {
					return result[i].GOOS < result[j].GOOS
				}
				return result[i].GOARCH < result[j].GOARCH
			})
			sort.Slice(d.expected, func(i, j int) bool {
				if d.expected[i].GOOS != d.expected[j].GOOS {
					return d.expected[i].GOOS < d.expected[j].GOOS
				}
				return d.expected[i].GOARCH < d.expected[j].GOARCH
			})

			if diff := cmp.Diff(d.expected, result); diff != "" {
				t.Fatalf("Unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
