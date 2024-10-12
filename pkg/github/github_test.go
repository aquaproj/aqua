package github_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if client := github.New(context.Background(), nil); client == nil {
		t.Fatal("client must not be nil")
	}
}
