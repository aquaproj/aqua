package versiongetter_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
)

var errUnreachable = errors.New("unreachable")

// fakeVersionsFile answers with a package's versions.json, or with err.
type fakeVersionsFile struct {
	versions []*g2.Version
	err      error
}

func (f *fakeVersionsFile) ReadVersionsFile(_ context.Context, _ *slog.Logger, pkgName string) (*g2.Versions, error) {
	if f.err != nil {
		return nil, fmt.Errorf("%w: %s", f.err, pkgName)
	}
	return &g2.Versions{Versions: f.versions}, nil
}

// fakeDirs resolves every name to one directory, or fails with err.
type fakeDirs struct{ err error }

func (f *fakeDirs) DirOf(_ context.Context, _ *slog.Logger, pkgName string) (string, error) {
	if f.err != nil {
		return "", fmt.Errorf("%w: %s", f.err, pkgName)
	}
	return g2.PackageDir("1790772767"), nil
}

// failingUpstream fails the test when it is asked: a project with a lock file must not
// ask upstream for a package of the standard registry.
type failingUpstream struct{ t *testing.T }

func (u *failingUpstream) Get(context.Context, *slog.Logger, *registry.PackageInfo, []*versiongetter.Filter) (string, error) {
	u.t.Error("upstream must not be asked")
	return "v9.9.9", nil
}

func (u *failingUpstream) List(context.Context, *slog.Logger, *registry.PackageInfo, []*versiongetter.Filter, int) ([]*fuzzyfinder.Item, error) {
	u.t.Error("upstream must not be asked")
	return fuzzyfinder.ConvertStringsToItems([]string{"v9.9.9"}), nil
}

func lockedGetter(t *testing.T, files *fakeVersionsFile, tree *fakeTreeGetter, dirs *fakeDirs) *versiongetter.FuzzyGetter {
	t.Helper()
	g := versiongetter.NewG2(g2.NewVersionLister(tree, dirs, "", ""), files)
	return versiongetter.NewFuzzy(fuzzyfinder.NewMock([]int{0}, nil), &failingUpstream{t: t}, g)
}

func versions(vs ...string) []*g2.Version {
	out := make([]*g2.Version, len(vs))
	for i, v := range vs {
		out[i] = &g2.Version{Version: v}
	}
	return out
}

func getLocked(t *testing.T, fg *versiongetter.FuzzyGetter, useFinder bool) (string, error) {
	t.Helper()
	return fg.Get(t.Context(), slog.New(slog.DiscardHandler), versiongetter.SourceG2, &registry.PackageInfo{Name: "cli/cli"}, "", useFinder, -1) //nolint:wrapcheck
}

// The version is chosen from versions.json, by version rather than by the order of the
// list, and the directory isn't read.
func TestFuzzyGetter_Get_g2Locked(t *testing.T) {
	t.Parallel()
	fg := lockedGetter(t, &fakeVersionsFile{versions: versions("v2.0.0", "v10.0.0", "v3.0.0")}, &fakeTreeGetter{versions: []string{"v99.0.0"}}, &fakeDirs{})
	got, err := getLocked(t, fg, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "v10.0.0" {
		t.Errorf("got %q, want v10.0.0", got)
	}
}

// A version versions.json marks as a prerelease isn't chosen, even when its tag doesn't
// say so. A package with nothing but prereleases gets the newest prerelease, as it did
// from the directory.
func TestFuzzyGetter_Get_g2LockedPrerelease(t *testing.T) {
	t.Parallel()
	vs := versions("v2.0.0", "v1.0.0")
	vs[0].Prerelease = true
	fg := lockedGetter(t, &fakeVersionsFile{versions: vs}, &fakeTreeGetter{}, &fakeDirs{})
	if got, err := getLocked(t, fg, false); err != nil || got != "v1.0.0" {
		t.Errorf("got %q, %v, want v1.0.0", got, err)
	}

	vs = versions("v2.0.0", "v1.0.0")
	vs[0].Prerelease = true
	vs[1].Prerelease = true
	fg = lockedGetter(t, &fakeVersionsFile{versions: vs}, &fakeTreeGetter{}, &fakeDirs{})
	if got, err := getLocked(t, fg, false); err != nil || got != "v2.0.0" {
		t.Errorf("got %q, %v, want v2.0.0", got, err)
	}
}

// versions.json that can't be read leaves the choice to the directory, which a package
// with no file yet has as well.
func TestFuzzyGetter_Get_g2LockedFallsBackToDirectory(t *testing.T) {
	t.Parallel()
	fg := lockedGetter(t, &fakeVersionsFile{err: errUnreachable}, &fakeTreeGetter{versions: []string{"v1.0.0", "v1.1.0"}}, &fakeDirs{})
	if got, err := getLocked(t, fg, false); err != nil || got != "v1.1.0" {
		t.Errorf("got %q, %v, want v1.1.0", got, err)
	}
}

// A package g2 doesn't hold, or holds no version of, is left alone without asking
// upstream, whether a version is chosen or listed.
func TestFuzzyGetter_Get_g2LockedNothing(t *testing.T) {
	t.Parallel()
	for name, fg := range map[string]*versiongetter.FuzzyGetter{
		"not held":       lockedGetter(t, &fakeVersionsFile{err: g2.ErrNoPackageBranch}, &fakeTreeGetter{notFound: true}, &fakeDirs{err: g2.ErrNoPackageBranch}),
		"no versions":    lockedGetter(t, &fakeVersionsFile{err: errUnreachable}, &fakeTreeGetter{notFound: true}, &fakeDirs{}),
		"empty versions": lockedGetter(t, &fakeVersionsFile{}, &fakeTreeGetter{}, &fakeDirs{}),
	} {
		for _, useFinder := range []bool{false, true} {
			got, err := getLocked(t, fg, useFinder)
			if err != nil || got != "" {
				t.Errorf("%s (finder %v): got %q, %v, want nothing", name, useFinder, got, err)
			}
		}
	}
}

// A g2 that can't be read fails the package rather than leaving it alone.
func TestFuzzyGetter_Get_g2LockedUnreadable(t *testing.T) {
	t.Parallel()
	fg := lockedGetter(t, &fakeVersionsFile{err: errUnreachable}, &fakeTreeGetter{}, &fakeDirs{err: errUnreachable})
	for _, useFinder := range []bool{false, true} {
		if _, err := getLocked(t, fg, useFinder); !errors.Is(err, errUnreachable) {
			t.Errorf("finder %v: got %v, want the failure to read g2", useFinder, err)
		}
	}
}

// A project without a lock file still asks upstream when g2 can't be read at all, the
// table of names included.
func TestFuzzyGetter_Get_g2OrUpstreamUnreadable(t *testing.T) {
	t.Parallel()
	upstream := versiongetter.NewMockVersionGetter(map[string][]*fuzzyfinder.Item{
		"cli/cli": fuzzyfinder.ConvertStringsToItems([]string{"v3.0.0"}),
	})
	g := versiongetter.NewG2(g2.NewVersionLister(&fakeTreeGetter{}, &fakeDirs{err: errUnreachable}, "", ""), nil)
	fg := versiongetter.NewFuzzy(fuzzyfinder.NewMock([]int{0}, nil), upstream, g)
	for _, useFinder := range []bool{false, true} {
		got, err := fg.Get(t.Context(), slog.New(slog.DiscardHandler), versiongetter.SourceG2OrUpstream, &registry.PackageInfo{Name: "cli/cli"}, "", useFinder, -1)
		if err != nil || got != "v3.0.0" {
			t.Errorf("finder %v: got %q, %v, want v3.0.0 from upstream", useFinder, got, err)
		}
	}
}
