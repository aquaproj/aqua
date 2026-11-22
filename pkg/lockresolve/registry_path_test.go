package lockresolve_test

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/github"
	installregistry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
)

// movingRef is a registry repository whose ref moves from commit to commit, as a moved tag
// or a branch does. It serves the registry file at every commit it has had, and resolves
// the ref to the commit it points at now.
type movingRef struct {
	head    string
	files   map[string]string
	offline bool
}

func newMovingRef() *movingRef {
	return &movingRef{files: map[string]string{}}
}

func (m *movingRef) DownloadGitHubContentFile(_ context.Context, _ *slog.Logger, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error) {
	if m.offline {
		return nil, errors.New("offline")
	}
	content, ok := m.files[param.Ref]
	if !ok {
		return nil, errors.New("not found")
	}
	return &domain.GitHubContentFile{String: content}, nil
}

func (m *movingRef) GetCommitSHA1(context.Context, string, string, string, string) (string, *github.Response, error) {
	return m.head, nil, nil
}

// moveTo points the ref at a new commit holding a registry with one package.
func (m *movingRef) moveTo(commit, repoName string) {
	m.head = commit
	m.files[commit] = "packages:\n- type: github_release\n  repo_owner: corp\n  repo_name: " + repoName + "\n"
}

func movingRefConfig() *aqua.Config {
	return &aqua.Config{
		Registries: aqua.Registries{
			"corp": {Name: "corp", Type: "github_content", RepoOwner: "corp", RepoName: "registry", Ref: "v1.0.0", Path: "registry.yaml"},
		},
	}
}

func newRealInstaller(rootDir string, ref *movingRef) *installregistry.Installer {
	return installregistry.New(&config.Param{RootDir: rootDir, MaxParallelism: 5}, ref, &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}, &cosign.MockVerifier{}, &slsa.MockVerifier{})
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

func noForce(*lockfile.Registry) bool { return false }

func acceptAll(*lockfile.Registry) bool { return true }

// record runs the lock update command's recording for the corp registry into lf.
func record(t *testing.T, inst *installregistry.Installer, ref *movingRef, lf *lockfile.LockFile) string {
	t.Helper()
	rgsts := lockresolve.NewRecordingRegistries(inst, ref, movingRefConfig(), "aqua.yaml", lf, noForce)
	got, err := readCorp(t, rgsts)
	if err != nil {
		t.Fatal(err)
	}
	rgsts.Commit(acceptAll)
	return got
}

// A lock file records the commit the ref pointed at, so what it installs doesn't change
// when the ref moves on: not on a machine that has never downloaded the registry, and not
// on one where another configuration recorded the new commit.
func TestRegistries_commitSurvivesMovedRef(t *testing.T) {
	t.Parallel()
	commit1 := strings.Repeat("1", 40)
	commit2 := strings.Repeat("2", 40)
	ref := newMovingRef()
	ref.moveTo(commit1, "old")

	shared := t.TempDir()
	inst := newRealInstaller(shared, ref)
	cfg := movingRefConfig()

	// Project A records the commit the ref points at first.
	lfA := lockfile.New()
	record(t, inst, ref, lfA)
	if r := lfA.FindRegistry(lockresolve.LockFileRegistry(cfg.Registries["corp"])); r == nil || r.Commit != commit1 {
		t.Fatalf("recorded %+v, want the commit the ref pointed at", r)
	}

	// The ref moves on.
	ref.moveTo(commit2, "new")

	// A machine that has never downloaded the registry installs what A recorded.
	if got, err := readCorp(t, lockresolve.NewRegistries(newRealInstaller(t.TempDir(), ref), cfg, filepath.Join(shared, "aqua.yaml"), lfA)); err != nil || got != "old" {
		t.Fatalf("a fresh machine read %q (%v), want what the lock file recorded", got, err)
	}

	// Project B records the new commit on the shared machine.
	if got := record(t, inst, ref, lockfile.New()); got != "new" {
		t.Fatalf("project B read %q, want the new content", got)
	}

	// Project A still reads what it recorded, from disk.
	ref.offline = true
	if got, err := readCorp(t, lockresolve.NewRegistries(inst, cfg, "aqua.yaml", lfA)); err != nil || got != "old" {
		t.Fatalf("project A read %q (%v), want what its lock file recorded", got, err)
	}
	ref.offline = false

	// A's lock update without --force keeps the recorded commit.
	before := *lfA.FindRegistry(lockresolve.LockFileRegistry(cfg.Registries["corp"]))
	if got := record(t, inst, ref, lfA); got != "old" {
		t.Fatalf("lock update read %q, want the recorded commit", got)
	}
	if after := lfA.FindRegistry(lockresolve.LockFileRegistry(cfg.Registries["corp"])); *after != before {
		t.Errorf("lock update without --force changed the record to %+v", after)
	}
}
