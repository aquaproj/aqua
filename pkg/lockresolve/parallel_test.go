package lockresolve_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
)

// countingInstaller answers every registry with the packages corp/tool0 to corp/tool<n-1>,
// or fails with err, after a pause long enough for others to ask meanwhile. It counts the
// installations.
type countingInstaller struct {
	n     int
	err   error
	calls atomic.Int32
}

func (i *countingInstaller) InstallRegistry(context.Context, *slog.Logger, *aqua.Registry, string, *checksum.Checksums) (*registry.Config, error) {
	return i.install()
}

func (i *countingInstaller) InstallLockedRegistry(context.Context, *slog.Logger, *aqua.Registry, string, *checksum.Checksum) (*registry.Config, error) {
	return i.install()
}

func (i *countingInstaller) install() (*registry.Config, error) {
	i.calls.Add(1)
	time.Sleep(20 * time.Millisecond)
	if i.err != nil {
		return nil, i.err
	}
	pkgs := make(registry.PackageInfos, i.n)
	for j := range i.n {
		pkgs[j] = &registry.PackageInfo{
			Name:          fmt.Sprintf("corp/tool%d", j),
			Type:          "github_release",
			RepoOwner:     "corp",
			RepoName:      fmt.Sprintf("tool%d", j),
			Asset:         "tool_{{.OS}}_{{.Arch}}.tar.gz",
			Format:        "tar.gz",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
		}
	}
	return &registry.Config{PackageInfos: pkgs}, nil
}

// Packages of one registry are resolved in parallel by install. The registry is installed
// once, and resolving its packages at once is free of data races (run with -race).
func TestLocked_parallelSameRegistry(t *testing.T) {
	t.Parallel()
	const n = 8
	installer := &countingInstaller{n: n}
	lf := lockFileWithRegistry(privateRegistry(), "ccc")
	rgsts := lockresolve.NewRegistries(installer, standardConfig(), "aqua.yaml", lf)

	errs := make([]error, n)
	var wg sync.WaitGroup
	for j := range n {
		wg.Go(func() {
			name := fmt.Sprintf("corp/tool%d", j)
			entry := toolEntry()
			entry.Name = name
			entry.Commands = []string{fmt.Sprintf("tool%d", j)}
			def, err := locked(t, &fakeG2{}, rgsts, &aqua.Package{Name: name, Version: "v1.0.0", Registry: "private"}, entry)
			if err == nil && def.Name != name {
				err = fmt.Errorf("got the definition of %s", def.Name)
			}
			errs[j] = err
		})
	}
	wg.Wait()
	for j, err := range errs {
		if err != nil {
			t.Errorf("corp/tool%d: %v", j, err)
		}
	}
	if got := installer.calls.Load(); got != 1 {
		t.Errorf("installed the registry %d times, want once", got)
	}
}

// A registry that fails to install isn't tried again, in parallel or after: every package
// from it gets the same failure.
func TestRegistries_rememberFailure(t *testing.T) {
	t.Parallel()
	errOffline := errors.New("offline")
	installer := &countingInstaller{err: errOffline}
	lf := lockFileWithRegistry(privateRegistry(), "ccc")
	rgsts := lockresolve.NewRegistries(installer, standardConfig(), "aqua.yaml", lf)

	errs := make([]error, 4)
	var wg sync.WaitGroup
	for j := range 3 {
		wg.Go(func() {
			_, errs[j] = rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private")
		})
	}
	wg.Wait()
	_, errs[3] = rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private")
	for j, err := range errs {
		if !errors.Is(err, errOffline) {
			t.Errorf("call %d: got %v, want the installation's failure", j, err)
		}
	}
	if got := installer.calls.Load(); got != 1 {
		t.Errorf("installed the registry %d times, want once", got)
	}
}

// What the lock update command records is the same when packages read the registry at once.
func TestRegistries_recordParallel(t *testing.T) {
	t.Parallel()
	installer := toolInstaller()
	lf := &lockfile.LockFile{}
	rgsts := lockresolve.NewRecordingRegistries(installer, &fakeResolver{}, standardConfig(), "aqua.yaml", lf, noForce)
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if len(installer.downloaded) != 1 {
		t.Errorf("downloaded %v, want the registry once", installer.downloaded)
	}
	if !rgsts.Commit(acceptAll) {
		t.Fatal("the lock file must change")
	}
	assertRecorded(t, lf.FindRegistry(lockresolve.LockFileRegistry(privateRegistry())))
}
