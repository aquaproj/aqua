package lockresolve_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	installregistry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
)

// movingRef serves whatever the ref points at now, which a moved tag or a branch changes.
type movingRef struct {
	content string
	offline bool
}

func (m *movingRef) DownloadGitHubContentFile(context.Context, *slog.Logger, *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error) {
	if m.offline {
		return nil, errors.New("offline")
	}
	return &domain.GitHubContentFile{String: m.content}, nil
}

func registryServing(repoName string) string {
	return "packages:\n- type: github_release\n  repo_owner: corp\n  repo_name: " + repoName + "\n"
}

func movingRefConfig() *aqua.Config {
	return &aqua.Config{
		Registries: aqua.Registries{
			"corp": {Name: "corp", Type: "github_content", RepoOwner: "corp", RepoName: "registry", Ref: "v1.0.0", Path: "registry.yaml"},
		},
	}
}

func readCorp(t *testing.T, rgsts *lockresolve.Registries) (string, error) {
	t.Helper()
	content, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "corp")
	if err != nil {
		return "", err //nolint:wrapcheck
	}
	if len(content.PackageInfos) != 1 {
		t.Fatalf("got %d packages, want 1", len(content.PackageInfos))
	}
	return content.PackageInfos[0].RepoName, nil
}

// The registry file of a ref is shared by every configuration on the machine. Once the ref
// points at new content, recording it, as another configuration's lock update does or as a
// --force run whose packages fail does before its checksum is committed, must not change
// what a lock file that recorded the earlier content reads.
func TestRegistries_lockedFileSurvivesMovedRef(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	ref := &movingRef{content: registryServing("old")}
	inst := installregistry.New(&config.Param{RootDir: rootDir, MaxParallelism: 5}, ref, &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}, &cosign.MockVerifier{}, &slsa.MockVerifier{})
	cfg := movingRefConfig()
	cfgFilePath := filepath.Join(rootDir, "aqua.yaml")
	accept := func(*lockfile.Registry) bool { return true }

	// Project A records the content the ref points at first.
	lfA := lockfile.New()
	recA := lockresolve.NewRecordingRegistries(inst, cfg, cfgFilePath, lfA, func(*lockfile.Registry) bool { return false })
	if _, err := readCorp(t, recA); err != nil {
		t.Fatal(err)
	}
	if !recA.Commit(accept) {
		t.Fatal("the checksum must be recorded")
	}

	// The ref moves, and project B records the new content.
	ref.content = registryServing("new")
	recB := lockresolve.NewRecordingRegistries(inst, cfg, cfgFilePath, lockfile.New(), func(*lockfile.Registry) bool { return false })
	if got, err := readCorp(t, recB); err != nil || got != "new" {
		t.Fatalf("project B read %q (%v), want the new content", got, err)
	}

	// Project A still reads what it recorded, from disk.
	ref.offline = true
	if got, err := readCorp(t, lockresolve.NewRegistries(inst, cfg, cfgFilePath, lfA)); err != nil || got != "old" {
		t.Fatalf("project A read %q (%v), want what its lock file recorded", got, err)
	}

	// Without the file on disk, project A downloads what the ref points at now, which
	// doesn't match its checksum.
	if err := os.RemoveAll(filepath.Join(rootDir, "registries")); err != nil {
		t.Fatal(err)
	}
	ref.offline = false
	if _, err := readCorp(t, lockresolve.NewRegistries(inst, cfg, cfgFilePath, lfA)); err == nil {
		t.Fatal("content that doesn't match the lock file must be refused")
	}
}
