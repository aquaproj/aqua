package lockresolve_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
)

// refOrCommit says which way a registry was read: at its ref, or at the commit the lock
// file recorded.
type refOrCommit struct {
	atRef     []string
	atCommit  []string
	checksums *checksum.Checksums
}

func (i *refOrCommit) InstallRegistry(_ context.Context, _ *slog.Logger, rgst *aqua.Registry, _ string, checksums *checksum.Checksums) (*registry.Config, error) {
	i.atRef = append(i.atRef, rgst.Name)
	i.checksums = checksums
	return &registry.Config{}, nil
}

func (i *refOrCommit) InstallLockedRegistry(_ context.Context, _ *slog.Logger, rgst *aqua.Registry, commit string, _ *checksum.Checksum) (*registry.Config, error) {
	i.atCommit = append(i.atCommit, rgst.Name+"@"+commit)
	return &registry.Config{}, nil
}

func updateConfig() *aqua.Config {
	return &aqua.Config{Registries: aqua.Registries{
		"recorded":   {Name: "recorded", Type: "github_content", RepoOwner: "corp", RepoName: "a", Ref: "v1.0.0", Path: "registry.yaml"},
		"unrecorded": {Name: "unrecorded", Type: "github_content", RepoOwner: "corp", RepoName: "b", Ref: "v1.0.0", Path: "registry.yaml"},
	}}
}

// The update command reads a registry the lock file has a record of at the recorded
// commit, and one it has none of at its ref, checked against the checksums it was given.
func TestNewUpdateRegistries(t *testing.T) {
	t.Parallel()
	cfg := updateConfig()
	key := lockresolve.LockFileRegistry(cfg.Registries["recorded"])
	key.Commit = "abc"
	key.Checksum = "0000"
	key.ChecksumAlgorithm = "sha256"
	lf := &lockfile.LockFile{}
	lf.SetRegistry(key)
	installer := &refOrCommit{}
	chks := checksum.New()
	rgsts := lockresolve.NewUpdateRegistries(installer, cfg, "aqua.yaml", lf, chks)
	logger := slog.New(slog.DiscardHandler)
	for _, name := range []string{"recorded", "unrecorded"} {
		if _, err := rgsts.Get(t.Context(), logger, name); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	if len(installer.atCommit) != 1 || installer.atCommit[0] != "recorded@abc" {
		t.Errorf("read at a commit %v, want recorded@abc", installer.atCommit)
	}
	if len(installer.atRef) != 1 || installer.atRef[0] != "unrecorded" {
		t.Errorf("read at a ref %v, want unrecorded", installer.atRef)
	}
	if installer.checksums != chks {
		t.Error("the registry read at its ref must be checked against the checksums given")
	}
}

// Install still fails on a registry with no record.
func TestNewRegistries_unrecorded(t *testing.T) {
	t.Parallel()
	installer := &refOrCommit{}
	rgsts := lockresolve.NewRegistries(installer, updateConfig(), "aqua.yaml", &lockfile.LockFile{})
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "unrecorded"); !errors.Is(err, lockresolve.ErrRegistryNotInLockFile) {
		t.Errorf("got %v, want the registry not being in the lock file", err)
	}
	if len(installer.atRef) != 0 {
		t.Error("the registry must not be read at its ref")
	}
}
