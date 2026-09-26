package lockupdate_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/lockupdate"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/resolve"
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

// fakeRegistryInstaller answers with one package definition, whatever is asked for.
type fakeRegistryInstaller struct {
	pkgInfo *registry.PackageInfo
}

func (i *fakeRegistryInstaller) InstallRegistries(_ context.Context, _ *slog.Logger, cfg *aqua.Config, _ string, _ *checksum.Checksums) (map[string]*registry.Config, error) {
	contents := make(map[string]*registry.Config, len(cfg.Registries))
	for name := range cfg.Registries {
		contents[name] = &registry.Config{PackageInfos: registry.PackageInfos{i.pkgInfo}}
	}
	return contents, nil
}

// fakeChecksumGetter answers with one checksum for every environment, under the
// identifier the entry computes for itself.
type fakeChecksumGetter struct {
	checksum string
}

func (g *fakeChecksumGetter) GetVerified(_ context.Context, logger *slog.Logger, checksums *checksum.Checksums, pkg *config.Package, supportedEnvs []string) error {
	if g.checksum == "" {
		return nil
	}
	rts, err := resolve.Runtimes(pkg.PackageInfo, supportedEnvs)
	if err != nil {
		return fmt.Errorf("get the runtimes: %w", err)
	}
	for _, rt := range rts {
		info := pkg.PackageInfo.Copy()
		info.OverrideByRuntime(rt)
		p := &config.Package{Package: pkg.Package, PackageInfo: info}
		id, err := p.ChecksumID(rt)
		if err != nil {
			return fmt.Errorf("get a checksum id: %w", err)
		}
		checksums.Set(id, &checksum.Checksum{ID: id, Checksum: g.checksum, Algorithm: "sha256"})
	}
	return nil
}

func pkg(name, version string) *aqua.Package {
	return &aqua.Package{Name: name, Version: version, Registry: aqua.RegistryTypeStandard}
}

func run(t *testing.T, cfg *aqua.Config, resolver *fakeResolver, args *lockupdate.Args) (string, error) {
	t.Helper()
	dir := t.TempDir()
	cfgFilePath := filepath.Join(dir, "aqua.yaml")
	ctrl := lockupdate.New(&fakeFinder{paths: []string{cfgFilePath}}, &fakeReader{cfg: cfg}, nil, nil, resolver)
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
		nil, nil, resolver,
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
		nil, nil, resolver,
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

// A version the machine decides can't be locked, and failing the run over it would
// cost every other package its entry.
func TestController_Update_skipNoVersion(t *testing.T) {
	t.Parallel()
	resolver := &fakeResolver{}
	cfg := &aqua.Config{Packages: []*aqua.Package{
		{Name: "cli/cli", Registry: aqua.RegistryTypeStandard},
		pkg("suzuki-shunsuke/tfcmt", "v4.0.0"),
	}}
	if _, err := run(t, cfg, resolver, &lockupdate.Args{}); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"suzuki-shunsuke/tfcmt@v4.0.0"}, resolver.resolved); diff != "" {
		t.Errorf("the resolved packages are wrong (-want +got):\n%s", diff)
	}
}

// A package from a registry other than the standard one has no branch in
// aqua-registry-g2, so it is resolved from its own registry rather than asked for
// there.
func TestController_Update_otherRegistry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgFilePath := filepath.Join(dir, "aqua.yaml")
	cfg := &aqua.Config{
		Registries: aqua.Registries{
			"foo": {Name: "foo", Type: "github_content", RepoOwner: "suzuki-shunsuke", RepoName: "my-registry", Ref: "v1.0.0", Path: "registry.yaml"},
		},
		Packages: []*aqua.Package{{Name: "foo/foo", Version: "v1.0.0", Registry: "foo"}},
	}
	g2 := &fakeResolver{}
	ctrl := lockupdate.New(
		&fakeFinder{paths: []string{cfgFilePath}},
		&fakeReader{cfg: cfg},
		&fakeRegistryInstaller{pkgInfo: &registry.PackageInfo{
			Name:          "foo/foo",
			Type:          "github_release",
			RepoOwner:     "foo",
			RepoName:      "foo",
			Asset:         "foo_{{.OS}}_{{.Arch}}.tar.gz",
			Format:        "tar.gz",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
		}},
		&fakeChecksumGetter{checksum: "abc"},
		g2,
	)
	if err := ctrl.Update(t.Context(), slog.New(slog.DiscardHandler), &config.Param{}, &lockupdate.Args{}); err != nil {
		t.Fatal(err)
	}

	if len(g2.resolved) != 0 {
		t.Errorf("g2 was asked for %v, want nothing", g2.resolved)
	}
	lf, err := lockfile.ReadFile(filepath.Join(dir, lockfile.FileName))
	if err != nil {
		t.Fatal(err)
	}
	if len(lf.Packages) != 1 {
		t.Fatalf("got %d entries, want 1", len(lf.Packages))
	}
	got := lf.Packages[0]
	if got.Asset != "foo_linux_amd64.tar.gz" {
		t.Errorf("the asset is %q, want foo_linux_amd64.tar.gz", got.Asset)
	}
	// The checksum comes from the getter, not from the registry.
	if got.Checksum != "abc" {
		t.Errorf("the checksum is %q, want abc", got.Checksum)
	}
	// The entry records which registry answered.
	if got.Registry == nil || got.Registry.RepoName != "my-registry" {
		t.Errorf("the registry is %+v, want my-registry", got.Registry)
	}
}

// The lock file has to carry a checksum for anything aqua downloads, so a package
// whose checksum can't be found is an error rather than an entry without one.
func TestController_Update_otherRegistry_noChecksum(t *testing.T) {
	t.Parallel()
	cfg := &aqua.Config{
		Registries: aqua.Registries{"foo": {Name: "foo", Type: "github_content"}},
		Packages:   []*aqua.Package{{Name: "foo/foo", Version: "v1.0.0", Registry: "foo"}},
	}
	ctrl := lockupdate.New(
		&fakeFinder{paths: []string{filepath.Join(t.TempDir(), "aqua.yaml")}},
		&fakeReader{cfg: cfg},
		&fakeRegistryInstaller{pkgInfo: &registry.PackageInfo{
			Type:          "github_release",
			RepoOwner:     "foo",
			RepoName:      "foo",
			Asset:         "foo_{{.OS}}_{{.Arch}}.tar.gz",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
		}},
		&fakeChecksumGetter{},
		&fakeResolver{},
	)
	if err := ctrl.Update(t.Context(), slog.New(slog.DiscardHandler), &config.Param{}, &lockupdate.Args{}); err == nil {
		t.Fatal("an error must be returned")
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

// A definition that keeps everything in version_overrides has no asset name at the
// top level. The checksums have to be fetched from the definition for this version,
// not from the one the registry holds.
//
// suzuki-shunsuke/tfcmt is written that way — version_constraint "false" at the top
// and an asset in every override — and the fetch asked for an empty asset name,
// which failed with "the asset isn't found".
func TestController_Update_otherRegistry_versionOverrides(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := &aqua.Config{
		Registries: aqua.Registries{"foo": {Name: "foo", Type: "local", Path: "registry.yaml"}},
		Packages:   []*aqua.Package{{Name: "foo/foo", Version: "v1.0.0", Registry: "foo"}},
	}
	getter := &recordingChecksumGetter{inner: &fakeChecksumGetter{checksum: "abc"}}
	ctrl := lockupdate.New(
		&fakeFinder{paths: []string{filepath.Join(dir, "aqua.yaml")}},
		&fakeReader{cfg: cfg},
		&fakeRegistryInstaller{pkgInfo: &registry.PackageInfo{
			Name:               "foo/foo",
			Type:               "github_release",
			RepoOwner:          "foo",
			RepoName:           "foo",
			VersionConstraints: "false",
			VersionOverrides: []*registry.VersionOverride{
				{
					VersionConstraints: "true",
					Asset:              "foo_{{.OS}}_{{.Arch}}.tar.gz",
					Format:             "tar.gz",
					SupportedEnvs:      registry.SupportedEnvs{"linux/amd64"},
				},
			},
		}},
		getter,
		&fakeResolver{},
	)
	if err := ctrl.Update(t.Context(), slog.New(slog.DiscardHandler), &config.Param{}, &lockupdate.Args{}); err != nil {
		t.Fatal(err)
	}
	if getter.asset != "foo_{{.OS}}_{{.Arch}}.tar.gz" {
		t.Errorf("the checksums were fetched for asset %q, want the one the override defines", getter.asset)
	}
	lf, err := lockfile.ReadFile(filepath.Join(dir, lockfile.FileName))
	if err != nil {
		t.Fatal(err)
	}
	if len(lf.Packages) != 1 || lf.Packages[0].Asset != "foo_linux_amd64.tar.gz" {
		t.Errorf("the entries are %+v", lf.Packages)
	}
}

// recordingChecksumGetter remembers the definition it was asked about.
type recordingChecksumGetter struct {
	inner *fakeChecksumGetter
	asset string
}

func (g *recordingChecksumGetter) GetVerified(ctx context.Context, logger *slog.Logger, checksums *checksum.Checksums, pkg *config.Package, supportedEnvs []string) error {
	g.asset = pkg.PackageInfo.Asset
	return g.inner.GetVerified(ctx, logger, checksums, pkg, supportedEnvs)
}
