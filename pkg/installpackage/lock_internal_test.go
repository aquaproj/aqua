package installpackage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// fakeLockSource answers with the registry's definition of gh, whose only command is gh.
type fakeLockSource struct{}

func (fakeLockSource) Locked(_ context.Context, _ *slog.Logger, pkg *aqua.Package, entry *lockfile.Entry) (*lockfile.Package, error) {
	if len(entry.Commands) != 1 || entry.Commands[0] != "gh" {
		return nil, errors.New("the commands don't match")
	}
	return &lockfile.Package{
		Name: pkg.Name, Version: pkg.Version, OS: "linux", Arch: "amd64",
		Type: "github_release", RepoOwner: "cli", RepoName: "cli", Asset: "gh.tar.gz",
		Checksum: entry.Checksum, ChecksumAlgorithm: "sha256",
		Files: []*lockfile.File{{Name: "gh", Src: "bin/gh"}},
	}, nil
}

func newLockTestInstaller() *Installer {
	return &Installer{runtime: &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}, maxParallelism: 5}
}

func lockedTestTarget(commands ...string) *Target {
	entry := &lockfile.Entry{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Checksum: "aaa", Commands: commands}
	return &Target{
		Pkg: &config.Package{
			Package:     &aqua.Package{Name: "cli/cli", Version: "v2.1.0", Registry: "standard"},
			PackageInfo: CommandsPackageInfo("cli/cli", entry),
		},
		Locked: true,
		Entry:  entry,
	}
}

// Links are created from the targets this returns, so a locked target comes back with the
// registry's definition, and one whose commands the registry doesn't confirm doesn't come
// back at all.
func TestInstaller_resolveLockedTargets(t *testing.T) {
	t.Parallel()
	is := newLockTestInstaller()
	got, failed := is.resolveLockedTargets(t.Context(), slog.New(slog.DiscardHandler), []*Target{
		lockedTestTarget("gh"),
		lockedTestTarget("gh", "git"),
	}, &ParamInstallPackages{LockSource: fakeLockSource{}})
	if !failed {
		t.Error("a target the registry doesn't confirm is a failure")
	}
	if len(got) != 1 {
		t.Fatalf("got %d targets, want 1", len(got))
	}
	if got[0].Entry != nil || got[0].Pkg.PackageInfo.Asset != "gh.tar.gz" || got[0].Checksum == nil {
		t.Errorf("the target isn't the registry's definition: %+v", got[0].Pkg.PackageInfo)
	}
}

// A locked target the tags or the names leave out isn't read from its registry: the run
// doesn't need it, and a registry it can't reach mustn't fail the run.
func TestInstaller_resolveLockedTargets_filtered(t *testing.T) {
	t.Parallel()
	is := newLockTestInstaller()
	private := lockedTestTarget("gh", "git") // the source would fail on it
	private.Pkg.Package.Tags = []string{"private"}
	for name, param := range map[string]*ParamInstallPackages{
		"excluded tag": {ExcludedTags: map[string]struct{}{"private": {}}},
		"another name": {Names: map[string]struct{}{"fzf": {}}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			param.LockSource = fakeLockSource{}
			got, failed := is.resolveLockedTargets(t.Context(), slog.New(slog.DiscardHandler), []*Target{private}, param)
			if failed || len(got) != 0 {
				t.Errorf("got %d targets, failed=%v; want none, not failed", len(got), failed)
			}
		})
	}
}

// A named target whose registry can't be read is reported for that, not also as a name no
// package matches.
func TestInstaller_resolveLockedTargets_foundNames(t *testing.T) {
	t.Parallel()
	is := newLockTestInstaller()
	found := map[string]struct{}{}
	_, failed := is.resolveLockedTargets(t.Context(), slog.New(slog.DiscardHandler), []*Target{
		lockedTestTarget("gh", "git"), // the source fails on it
	}, &ParamInstallPackages{
		LockSource: fakeLockSource{},
		Names:      map[string]struct{}{"gh": {}},
		FoundNames: found,
	})
	if !failed {
		t.Error("a target that can't be read is a failure")
	}
	if _, ok := found["gh"]; !ok {
		t.Error("the name the target matches must be recorded as found")
	}
}

// slowLockSource answers after a pause, slower for the earlier packages, and fails on the
// packages named in fail. It counts how many it answers at once.
type slowLockSource struct {
	fail    map[string]bool
	mu      sync.Mutex
	running int
	peak    int
}

func (s *slowLockSource) Locked(_ context.Context, _ *slog.Logger, pkg *aqua.Package, entry *lockfile.Entry) (*lockfile.Package, error) {
	s.mu.Lock()
	s.running++
	s.peak = max(s.peak, s.running)
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.running--
		s.mu.Unlock()
	}()
	var n int
	if _, err := fmt.Sscanf(pkg.Name, "o/p%d", &n); err != nil {
		return nil, err //nolint:wrapcheck
	}
	time.Sleep(time.Duration(10-n) * 5 * time.Millisecond)
	if s.fail[pkg.Name] {
		return nil, errors.New("the registry can't be read")
	}
	return &lockfile.Package{
		Name: pkg.Name, Version: pkg.Version, OS: "linux", Arch: "amd64",
		Type: "github_release", RepoOwner: "o", RepoName: pkg.Name[2:], Asset: "a.tar.gz",
		Checksum: entry.Checksum, ChecksumAlgorithm: "sha256",
	}, nil
}

func numberedTarget(n int) *Target {
	name := fmt.Sprintf("o/p%d", n)
	entry := &lockfile.Entry{Name: name, Version: "v1.0.0", OS: "linux", Arch: "amd64", Checksum: "aaa", Commands: []string{"p"}}
	return &Target{
		Pkg: &config.Package{
			Package:     &aqua.Package{Name: name, Version: "v1.0.0", Registry: "standard"},
			PackageInfo: CommandsPackageInfo(name, entry),
		},
		Locked: true,
		Entry:  entry,
	}
}

// The targets are read in parallel, up to the parallelism, and come back in the order they
// were given whichever is read first. A target that can't be read drops out alone, and the
// failures are logged in that order too.
func TestInstaller_resolveLockedTargets_parallel(t *testing.T) {
	t.Parallel()
	is := newLockTestInstaller()
	is.maxParallelism = 3
	src := &slowLockSource{fail: map[string]bool{"o/p2": true, "o/p5": true}}
	unlocked := &Target{Pkg: &config.Package{Package: &aqua.Package{Name: "o/unlocked"}}}
	targets := make([]*Target, 0, 9)
	targets = append(targets, unlocked)
	for i := range 8 {
		targets = append(targets, numberedTarget(i))
	}
	buf := &bytes.Buffer{}
	got, failed := is.resolveLockedTargets(t.Context(), slog.New(slog.NewTextHandler(buf, nil)), targets, &ParamInstallPackages{LockSource: src})
	if !failed {
		t.Error("a target that can't be read is a failure")
	}
	names := make([]string, len(got))
	for i, tg := range got {
		names[i] = tg.Pkg.Package.Name
	}
	if want := "o/unlocked o/p0 o/p1 o/p3 o/p4 o/p6 o/p7"; strings.Join(names, " ") != want {
		t.Errorf("got %s, want %s", strings.Join(names, " "), want)
	}
	if src.peak < 2 || src.peak > 3 {
		t.Errorf("read %d targets at once, want 2 or 3", src.peak)
	}
	logs := buf.String()
	if i, j := strings.Index(logs, "o/p2"), strings.Index(logs, "o/p5"); i < 0 || j < 0 || i > j {
		t.Errorf("the failures aren't logged in order:\n%s", logs)
	}
}
