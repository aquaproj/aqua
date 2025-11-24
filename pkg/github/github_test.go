package github_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if client := github.New(t.Context(), logrus.NewEntry(logrus.New())); client == nil {
		t.Fatal("client must not be nil")
	}
}

func Test_getGHESKeyName(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		want    string
		wantErr bool
	}{
		{
			name:   "Normal case",
			domain: "ghe.example.com",
			want:   "GITHUB_TOKEN_ghe_example_com",
		},
		{
			name:   "Domain with mixed case",
			domain: "gHe.eXaMpLe.CoM",
			want:   "GITHUB_TOKEN_ghe_example_com",
		},
		{
			name:   "Domain with hyphen",
			domain: "ghe-example.com", // only hyphen is allowed as special characters
			want:   "GITHUB_TOKEN_ghe-example_com",
		},
		{
			name:   "Domain with spaces",
			domain: " ghe.example.com ",
			want:   "GITHUB_TOKEN_ghe_example_com",
		},
		{
			name:   "handle github.com",
			domain: "github.com",
			want:   "GITHUB_TOKEN",
		},
		{
			name:    "invalid domain with underscore",
			domain:  "ghe_example_com", // underscore is not allowed in domains
			wantErr: true,
		},
		{
			name:    "invalid domain with slash",
			domain:  "github.com/",
			wantErr: true,
		},
		{
			name:    "invalid domain with scheme",
			domain:  "https://github.com",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := github.GetGHESTokenEnvKey(tt.domain)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, got)
			assert.NoError(t, err)
		})
	}
}
