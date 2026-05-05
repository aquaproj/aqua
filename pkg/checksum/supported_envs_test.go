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
			pkgSupportedEnvs:  []string{osLinuxAmd, osDarwinAmd, osWindowsAmd},
			confSupportedEnvs: []string{osLinux, osDarwin},
			expected: []*runtime.Runtime{
				{GOOS: osLinux, GOARCH: archAmd64},
				{GOOS: osDarwin, GOARCH: archAmd64},
			},
		},
		{
			name:              "no matching environments",
			pkgSupportedEnvs:  []string{osLinuxArm, osDarwinArm},
			confSupportedEnvs: []string{"windows"},
			expected:          []*runtime.Runtime{},
		},
		{
			name:              "empty package environments",
			pkgSupportedEnvs:  nil,
			confSupportedEnvs: []string{osLinux, osDarwin},
			expected: []*runtime.Runtime{
				{GOOS: osLinux, GOARCH: archAmd64},
				{GOOS: osLinux, GOARCH: archArm64},
				{GOOS: osDarwin, GOARCH: archAmd64},
				{GOOS: osDarwin, GOARCH: archArm64},
			},
		},
		{
			name:              "empty config environments returns all package envs",
			pkgSupportedEnvs:  []string{osLinuxAmd, osDarwinAmd},
			confSupportedEnvs: nil,
			expected: []*runtime.Runtime{
				{GOOS: osLinux, GOARCH: archAmd64},
				{GOOS: osDarwin, GOARCH: archAmd64},
			},
		},
		{
			name:              "partial matches",
			pkgSupportedEnvs:  []string{osLinuxAmd, osLinuxArm, osWindowsAmd},
			confSupportedEnvs: []string{osLinux},
			expected: []*runtime.Runtime{
				{GOOS: osLinux, GOARCH: archAmd64},
				{GOOS: osLinux, GOARCH: archArm64},
			},
		},
		{
			name:              "exact arch match",
			pkgSupportedEnvs:  []string{osLinuxAmd, osDarwinArm, osWindowsAmd},
			confSupportedEnvs: []string{archAmd64},
			expected: []*runtime.Runtime{
				{GOOS: osLinux, GOARCH: archAmd64},
				{GOOS: "windows", GOARCH: archAmd64},
			},
		},
		{
			name:              "mixed OS and arch filters",
			pkgSupportedEnvs:  []string{osLinuxAmd, osLinuxArm, osDarwinAmd, osDarwinArm, osWindowsAmd},
			confSupportedEnvs: []string{osLinux, archArm64},
			expected: []*runtime.Runtime{
				{GOOS: osLinux, GOARCH: archAmd64},
				{GOOS: osLinux, GOARCH: archArm64},
				{GOOS: osDarwin, GOARCH: archArm64},
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
