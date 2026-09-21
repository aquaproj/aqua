package lockupdate_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/controller/lockupdate"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/google/go-cmp/cmp"
)

type fakeFinder struct {
	paths []string
}

func (f *fakeFinder) Finds(_, _ string) []string { return f.paths }

type fakeReader struct {
	cfg *aqua.Config
}

func (r *fakeReader) Read(_ *slog.Logger, _ string, cfg *aqua.Config) error {
	*cfg = *r.cfg
	return nil
}

type fakeResolver struct {
	// resolved records every package version the controller asked for, which is how
	// the tests tell "left alone" apart from "resolved to the same thing".
	resolved []string
	err      error
}

func (r *fakeResolver) Resolve(_ context.Context, _ *slog.Logger, pkgName, version string) ([]*lockfile.Package, error) {
	r.resolved = append(r.resolved, pkgName+"@"+version)
	if r.err != nil {
		return nil, r.err
	}
	return []*lockfile.Package{
		{Name: pkgName, Version: version, OS: "linux", Arch: "amd64", Type: "github_release", Checksum: "x"},
		{Name: pkgName, Version: version, OS: "darwin", Arch: "arm64", Type: "github_release", Checksum: "y"},
	}, nil
}

func pkg(name, version string) *aqua.Package {
	return &aqua.Package{Name: name, Version: version, Registry: aqua.RegistryTypeStandard}
}

func run(t *testing.T, cfg *aqua.Config, resolver *fakeResolver, args *lockupdate.Args) (string, error) {
	t.Helper()
	dir := t.TempDir()
	cfgFilePath := filepath.Join(dir, "aqua.yaml")
	ctrl := lockupdate.New(&fakeFinder{paths: []string{cfgFilePath}}, &fakeReader{cfg: cfg}, resolver)
	err := ctrl.Update(t.Context(), slog.New(slog.DiscardHandler), &config.Param{}, args)
	return filepath.Join(dir, lockfile.FileName), err
}

func TestController_Update(t *testing.T) {
	t.Parallel()
	resolver := &fakeResolver{}
	cfg := &aqua.Config{Packages: []*aqua.Package{pkg("cli/cli", "v2.1.0")}}

	lockFilePath, err := run(t, cfg, resolver, &lockupdate.Args{})
	if err != nil {
		t.Fatal(err)
	}

	lf, err := lockfile.ReadFile(lockFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(lf.Packages) != 2 {
		t.Fatalf("got %d entries, want 2", len(lf.Packages))
	}
	if lf.SchemaVersion != lockfile.SchemaVersion {
		t.Errorf("schema version is %q, want %q", lf.SchemaVersion, lockfile.SchemaVersion)
	}
}

// A package already in the lock file keeps its entry. Re-resolving it would make the
// recorded answer depend on when the command last ran, which is what the file exists
// to prevent.
func TestController_Update_alreadyLocked(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgFilePath := filepath.Join(dir, "aqua.yaml")
	lockFilePath := filepath.Join(dir, lockfile.FileName)

	lf := lockfile.New()
	lf.Packages = []*lockfile.Package{
		{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Type: "github_release", Checksum: "old"},
	}
	if err := lockfile.Write(lockFilePath, lf); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(lockFilePath)
	if err != nil {
		t.Fatal(err)
	}

	resolver := &fakeResolver{}
	ctrl := lockupdate.New(
		&fakeFinder{paths: []string{cfgFilePath}},
		&fakeReader{cfg: &aqua.Config{Packages: []*aqua.Package{pkg("cli/cli", "v2.1.0")}}},
		resolver,
	)
	if err := ctrl.Update(t.Context(), slog.New(slog.DiscardHandler), &config.Param{}, &lockupdate.Args{}); err != nil {
		t.Fatal(err)
	}

	if len(resolver.resolved) != 0 {
		t.Errorf("resolved %v, want nothing", resolver.resolved)
	}
	after, err := os.ReadFile(lockFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(string(before), string(after)); diff != "" {
		t.Errorf("the lock file changed (-before +after):\n%s", diff)
	}
}

// --force replaces the entries rather than adding to them: the new resolution may
// cover different environments, and an entry left behind for a dropped one would be
// installed from.
func TestController_Update_force(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgFilePath := filepath.Join(dir, "aqua.yaml")
	lockFilePath := filepath.Join(dir, lockfile.FileName)

	lf := lockfile.New()
	lf.Packages = []*lockfile.Package{
		{Name: "cli/cli", Version: "v2.1.0", OS: "windows", Arch: "amd64", Type: "github_release", Checksum: "old"},
	}
	if err := lockfile.Write(lockFilePath, lf); err != nil {
		t.Fatal(err)
	}

	resolver := &fakeResolver{}
	ctrl := lockupdate.New(
		&fakeFinder{paths: []string{cfgFilePath}},
		&fakeReader{cfg: &aqua.Config{Packages: []*aqua.Package{pkg("cli/cli", "v2.1.0")}}},
		resolver,
	)
	if err := ctrl.Update(t.Context(), slog.New(slog.DiscardHandler), &config.Param{}, &lockupdate.Args{Force: true}); err != nil {
		t.Fatal(err)
	}

	got, err := lockfile.ReadFile(lockFilePath)
	if err != nil {
		t.Fatal(err)
	}
	oses := make([]string, 0, len(got.Packages))
	for _, p := range got.Packages {
		oses = append(oses, p.OS)
	}
	if diff := cmp.Diff([]string{"darwin", "linux"}, oses); diff != "" {
		t.Errorf("the environments are wrong (-want +got):\n%s", diff)
	}
}

func TestController_Update_selectPackages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "every package",
			want: []string{"cli/cli@v2.1.0", "suzuki-shunsuke/tfcmt@v4.0.0"},
		},
		{
			name: "by name",
			args: []string{"cli/cli"},
			want: []string{"cli/cli@v2.1.0"},
		},
		{
			// The version is a filter, not an instruction: a name@version that
			// doesn't match aqua.yaml selects nothing rather than locking a version
			// aqua.yaml doesn't ask for.
			name: "by name and version",
			args: []string{"cli/cli@v1.0.0"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resolver := &fakeResolver{}
			cfg := &aqua.Config{Packages: []*aqua.Package{
				pkg("cli/cli", "v2.1.0"),
				pkg("suzuki-shunsuke/tfcmt", "v4.0.0"),
			}}
			if _, err := run(t, cfg, resolver, &lockupdate.Args{Packages: tt.args}); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, resolver.resolved); diff != "" {
				t.Errorf("the resolved packages are wrong (-want +got):\n%s", diff)
			}
		})
	}
}

// A package aqua can't resolve yet is skipped, not failed on: aqua.yaml is allowed to
// hold packages g2 doesn't cover, and refusing to write the file would cost every
// other package its entry.
func TestController_Update_skip(t *testing.T) {
	t.Parallel()
	resolver := &fakeResolver{}
	cfg := &aqua.Config{Packages: []*aqua.Package{
		{Name: "cli/cli", Registry: aqua.RegistryTypeStandard},
		{Name: "foo/foo", Version: "v1.0.0", Registry: "local"},
		pkg("suzuki-shunsuke/tfcmt", "v4.0.0"),
	}}
	if _, err := run(t, cfg, resolver, &lockupdate.Args{}); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"suzuki-shunsuke/tfcmt@v4.0.0"}, resolver.resolved); diff != "" {
		t.Errorf("the resolved packages are wrong (-want +got):\n%s", diff)
	}
}

// A failure on one package must not cost the others their entries, so the run finishes
// and writes what it resolved before reporting.
func TestController_Update_resolveError(t *testing.T) {
	t.Parallel()
	resolver := &fakeResolver{err: errors.New("no such version")}
	cfg := &aqua.Config{Packages: []*aqua.Package{pkg("cli/cli", "v2.1.0")}}
	lockFilePath, err := run(t, cfg, resolver, &lockupdate.Args{})
	if err == nil {
		t.Fatal("an error must be returned")
	}
	if _, err := os.Stat(lockFilePath); !os.IsNotExist(err) {
		t.Error("the lock file must not be written when nothing was resolved")
	}
}

// Nothing to do must not rewrite the file. A command that always touched it would show
// up as a change in git even when the answer is the same.
func TestController_Update_noPackage(t *testing.T) {
	t.Parallel()
	lockFilePath, err := run(t, &aqua.Config{}, &fakeResolver{}, &lockupdate.Args{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(lockFilePath); !os.IsNotExist(err) {
		t.Error("the lock file must not be created when there is no package")
	}
}
