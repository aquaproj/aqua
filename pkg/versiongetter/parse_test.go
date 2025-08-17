package versiongetter_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-version"
)

func TestGetVersionAndPrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		tag           string
		wantVersion   string
		wantPrefix    string
		wantErr       bool
		wantNilResult bool
	}{
		{
			name:        "simple semantic version",
			tag:         "1.2.3",
			wantVersion: "1.2.3",
			wantPrefix:  "",
			wantErr:     false,
		},
		{
			name:        "version with v prefix",
			tag:         "v1.2.3",
			wantVersion: "1.2.3",
			wantPrefix:  "",
			wantErr:     false,
		},
		{
			name:        "version with custom prefix",
			tag:         "release-1.2.3",
			wantVersion: "1.2.3",
			wantPrefix:  "release-",
			wantErr:     false,
		},
		{
			name:        "version with custom prefix and v",
			tag:         "release-v1.2.3",
			wantVersion: "1.2.3",
			wantPrefix:  "release-",
			wantErr:     false,
		},
		{
			name:        "version with prerelease",
			tag:         "v1.2.3-rc1",
			wantVersion: "1.2.3-rc1",
			wantPrefix:  "",
			wantErr:     false,
		},
		{
			name:        "version with build metadata",
			tag:         "v1.2.3+build123",
			wantVersion: "1.2.3+build123",
			wantPrefix:  "",
			wantErr:     false,
		},
		{
			name:        "version with hyphen separator",
			tag:         "v1.2.3-alpha.1",
			wantVersion: "1.2.3-alpha.1",
			wantPrefix:  "",
			wantErr:     false,
		},
		{
			name:        "two digit version",
			tag:         "v1.2",
			wantVersion: "1.2.0",
			wantPrefix:  "",
			wantErr:     false,
		},
		{
			name:        "single digit version",
			tag:         "v1",
			wantVersion: "1.0.0",
			wantPrefix:  "",
			wantErr:     false,
		},
		{
			name:        "prefix with underscore",
			tag:         "my_release_v1.2.3",
			wantVersion: "1.2.3",
			wantPrefix:  "my_release_",
			wantErr:     false,
		},
		{
			name:        "complex prefix",
			tag:         "project-2024-release-v2.0.0",
			wantVersion: "2024.0.0-release-v2.0.0",
			wantPrefix:  "project-",
			wantErr:     false,
		},
		{
			name:        "version with dot separator in prerelease",
			tag:         "v1.0.0.beta1",
			wantVersion: "1.0.0.beta1",
			wantPrefix:  "",
			wantErr:     true,
		},
		{
			name:          "invalid version - no numbers",
			tag:           "not-a-version",
			wantNilResult: true,
			wantErr:       false,
		},
		{
			name:          "empty string",
			tag:           "",
			wantNilResult: true,
			wantErr:       false,
		},
		{
			name:        "version with leading zeros",
			tag:         "v01.02.03",
			wantVersion: "1.2.3",
			wantPrefix:  "",
			wantErr:     false,
		},
		{
			name:        "date-like prefix with version",
			tag:         "2024-01-15-v1.2.3",
			wantVersion: "2024.0.0-01-15-v1.2.3",
			wantPrefix:  "",
			wantErr:     false,
		},
		{
			name:        "kubernetes style version",
			tag:         "kubernetes-1.28.0",
			wantVersion: "1.28.0",
			wantPrefix:  "kubernetes-",
			wantErr:     false,
		},
		{
			name:        "version at the end without v",
			tag:         "release1.0.0",
			wantVersion: "1.0.0",
			wantPrefix:  "release",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotVersion, gotPrefix, err := versiongetter.GetVersionAndPrefix(tt.tag)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetVersionAndPrefix() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetVersionAndPrefix() unexpected error = %v", err)
				return
			}

			if tt.wantNilResult {
				if gotVersion != nil {
					t.Errorf("GetVersionAndPrefix() gotVersion = %v, want nil", gotVersion)
				}
				if gotPrefix != "" {
					t.Errorf("GetVersionAndPrefix() gotPrefix = %v, want empty", gotPrefix)
				}
				return
			}

			if gotVersion == nil {
				t.Errorf("GetVersionAndPrefix() gotVersion = nil, want %s", tt.wantVersion)
				return
			}

			wantVer, err := version.NewVersion(tt.wantVersion)
			if err != nil {
				t.Fatalf("Failed to create expected version: %v", err)
			}

			if !gotVersion.Equal(wantVer) {
				t.Errorf("GetVersionAndPrefix() gotVersion = %v, want %v", gotVersion, wantVer)
			}

			if diff := cmp.Diff(tt.wantPrefix, gotPrefix); diff != "" {
				t.Errorf("GetVersionAndPrefix() prefix mismatch (-want +got):\n%s", diff)
			}
		})
	}
}