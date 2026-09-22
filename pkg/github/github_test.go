package github_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

func TestNew(t *testing.T) {
	t.Parallel()
	client, err := github.New(t.Context(), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("client must not be nil")
	}
}
