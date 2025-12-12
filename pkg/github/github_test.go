package github_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if client := github.New(t.Context(), logrus.NewEntry(logrus.New())); client == nil {
		t.Fatal("client must not be nil")
	}
}

func Test_getGHESKeyName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		baseURL string
		want    string
		wantErr bool
	}{
		{
			name:    "Normal case",
			baseURL: "https://ghe.example.com",
			want:    "GITHUB_TOKEN_ghe_example_com",
		},
		{
			name:    "Resolve github.com for empty baseURL",
			baseURL: "",
			want:    "GITHUB_TOKEN",
		},
		{
			name:    "Domain with mixed case",
			baseURL: "https://gHe.eXaMpLe.CoM",
			want:    "GITHUB_TOKEN_ghe_example_com",
		},
		{
			name:    "Domain with hyphen",
			baseURL: "https://ghe-example.com", // only hyphen is allowed as special characters
			want:    "GITHUB_TOKEN_ghe-example_com",
		},
		{
			name:    "handle github.com",
			baseURL: "https://github.com",
			want:    "GITHUB_TOKEN",
		},
		{
			name:    "invalid domain with underscore",
			baseURL: "https://ghe_example_com.com", // underscore is not allowed in domains
			wantErr: true,
		},
		{
			name:    "non HTTPS scheme",
			baseURL: "http://ghe.example.com",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := github.GetGitHubTokenEnvKey(tt.baseURL)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
