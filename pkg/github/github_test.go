package github_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/pkg/github"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if client := github.New(context.Background()); client == nil {
		t.Fatal("client must not be nil")
	}
}
