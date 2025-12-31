package github_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if client := github.New(t.Context(), slog.New(slog.DiscardHandler)); client == nil {
		t.Fatal("client must not be nil")
	}
}
