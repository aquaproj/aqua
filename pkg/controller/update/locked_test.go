package update_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/update"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/goccy/go-yaml"
)

var errFake = errors.New("fake failure")

// lockedInstaller serves registry files by name, and remembers how each was read.
type lockedInstaller struct {
	mu           sync.Mutex
	contents     map[string]string // registry name -> registry.yaml
	unreadable   map[string]bool
	installedAll int
	atRef        []string
	atCommit     []string
}

func (i *lockedInstaller) InstallRegistries(context.Context, *slog.Logger, *aqua.Config, string, *checksum.Checksums) (map[string]*registry.Config, error) {
	i.installedAll++
	return nil, errFake
}

func (i *lockedInstaller) InstallRegistry(_ context.Context, _ *slog.Logger, rgst *aqua.Registry, _ string, _ *checksum.Checksums) (*registry.Config, error) {
	i.mu.Lock()
	i.atRef = append(i.atRef, rgst.Name)
	i.mu.Unlock()
	return i.read(rgst.Name)
}

func (i *lockedInstaller) InstallLockedRegistry(_ context.Context, _ *slog.Logger, rgst *aqua.Registry, commit string, _ *checksum.Checksum) (*registry.Config, error) {
	i.mu.Lock()
	i.atCommit = append(i.atCommit, rgst.Name+"@"+commit)
	i.mu.Unlock()
	return i.read(rgst.Name)
}

func (i *lockedInstaller) read(name string) (*registry.Config, error) {
	if i.unreadable[name] {
		return nil, errFake
	}
	cfg := &registry.Config{}
	if err := yaml.Unmarshal([]byte(i.contents[name]), cfg); err != nil {
		return nil, err //nolint:wrapcheck
	}
	return cfg, nil
}

type getterCall struct {
	source versiongetter.Source
	pkg    *registry.PackageInfo
}

// recordingGetter answers with a version per package name, and remembers what it was asked
// with.
type recordingGetter struct {
	versions map[string]string
	errs     map[string]bool
	calls    map[string]getterCall
}

func (g *recordingGetter) Get(_ context.Context, _ *slog.Logger, source versiongetter.Source, pkg *registry.PackageInfo, _ string, _ bool, _ int) (string, error) {
	if g.calls == nil {
		g.calls = map[string]getterCall{}
	}
	g.calls[pkg.GetName()] = getterCall{source: source, pkg: pkg}
	if g.errs[pkg.GetName()] {
		return "", errFake
	}
	return g.versions[pkg.GetName()], nil
}

// lockedGH counts how often a registry's latest release was asked for.
type lockedGH struct {
	github.MockRepositoriesService

	latest int
}

func (g *lockedGH) GetLatestRelease(context.Context, string, string) (*github.RepositoryRelease, *github.Response, error) {
	g.latest++
	tag := "v9.0.0"
	return &github.RepositoryRelease{TagName: tag}, nil, nil
}

const (
	corpRegistry = `
- name: corp
  type: github_content
  repo_owner: corp
  repo_name: registry
  ref: v1.0.0
  path: registry.yaml
`
	// toolDefinition is a package of a monorepo whose versions are the tags starting
	// with tool/, and which supports only darwin.
	toolDefinition = `packages:
  - type: github_release
    repo_owner: example-org
    repo_name: monorepo
    name: example-org/tool
    asset: tool.tar.gz
    version_prefix: tool/
    supported_envs: [darwin]
  - type: github_release
    repo_owner: example-org
    repo_name: vars
    name: example-org/vars
    asset: vars.tar.gz
    vars:
      - name: flavor
        required: true
`
)

type lockedHarness struct {
	dir       string
	param     *config.Param
	installer *lockedInstaller
	getter    *recordingGetter
	gh        *lockedGH
	which     *which.MockMultiController
	idxs      []int
}

func newLockedHarness(t *testing.T, aquaYAML string, lock any) *lockedHarness {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "aqua.yaml"), aquaYAML)
	lockPath := filepath.Join(dir, lockfile.FileName)
	switch lf := lock.(type) {
	case *lockfile.LockFile:
		if err := lockfile.Write(lockPath, lf); err != nil {
			t.Fatal(err)
		}
	case string:
		writeFile(t, lockPath, lf)
	}
	return &lockedHarness{
		dir:       dir,
		param:     &config.Param{CWD: dir, OnlyPackage: true},
		installer: &lockedInstaller{contents: map[string]string{"corp": toolDefinition}, unreadable: map[string]bool{}},
		getter:    &recordingGetter{versions: map[string]string{}, errs: map[string]bool{}},
		gh:        &lockedGH{},
		which:     &which.MockMultiController{},
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func (h *lockedHarness) run(t *testing.T) error {
	t.Helper()
	ctrl := update.New(h.param, h.gh, finder.NewConfigFinder(), reader.New(h.param), h.installer,
		&runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}, h.getter, fuzzyfinder.NewMock(h.idxs, nil), h.which)
	return ctrl.Update(t.Context(), slog.New(slog.DiscardHandler), h.param) //nolint:wrapcheck
}

func (h *lockedHarness) read(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(h.dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// recordCorp records the corp registry in the lock file at commit abc.
func recordCorp(t *testing.T) *lockfile.LockFile {
	t.Helper()
	cfg := &aqua.Config{}
	if err := yaml.Unmarshal([]byte("registries:"+corpRegistry), cfg); err != nil {
		t.Fatal(err)
	}
	key := lockresolve.LockFileRegistry(cfg.Registries["corp"])
	key.Commit = "abc"
	key.Checksum = strings.Repeat("0", 64)
	key.ChecksumAlgorithm = "sha256"
	lf := &lockfile.LockFile{}
	lf.SetRegistry(key)
	return lf
}

// A package of the standard registry is updated from aqua-registry-g2 alone: no registry
// is read, so one aqua-registry v1 at aqua.yaml's ref doesn't have can be updated, and the
// lock file needn't have it.
func TestController_Update_lockedStandard(t *testing.T) {
	t.Parallel()
	h := newLockedHarness(t, "registries:\n- type: standard\n  ref: v4.0.0\npackages:\n- name: cli/cli@v2.0.0\n", &lockfile.LockFile{})
	h.getter.versions["cli/cli"] = "v2.1.0"
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if got := h.read(t, "aqua.yaml"); !strings.Contains(got, "cli/cli@v2.1.0") {
		t.Errorf("aqua.yaml isn't updated:\n%s", got)
	}
	call := h.getter.calls["cli/cli"]
	if call.source != versiongetter.SourceG2 {
		t.Errorf("source %v, want SourceG2", call.source)
	}
	if h.installer.installedAll != 0 || len(h.installer.atRef) != 0 || len(h.installer.atCommit) != 0 {
		t.Errorf("a registry was read: all %d, at a ref %v, at a commit %v", h.installer.installedAll, h.installer.atRef, h.installer.atCommit)
	}
}

// A registry no package to update comes from isn't read, so one that can't be read doesn't
// stop the others.
func TestController_Update_lockedUnreadableRegistry(t *testing.T) {
	t.Parallel()
	h := newLockedHarness(t, `registries:
- type: standard
  ref: v4.300.0
- name: private
  type: github_content
  repo_owner: example-org
  repo_name: private-registry
  ref: v1.0.0
  path: registry.yaml
packages:
- name: cli/cli@v2.60.0
`, &lockfile.LockFile{})
	h.installer.unreadable["private"] = true
	h.getter.versions["cli/cli"] = "v2.61.0"
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if got := h.read(t, "aqua.yaml"); !strings.Contains(got, "cli/cli@v2.61.0") {
		t.Errorf("aqua.yaml isn't updated:\n%s", got)
	}
	if len(h.installer.atRef)+len(h.installer.atCommit) != 0 {
		t.Errorf("a registry was read: at a ref %v, at a commit %v", h.installer.atRef, h.installer.atCommit)
	}
}

// A package of another registry is looked for with its whole definition, read at the
// commit the lock file recorded, whatever environment it supports.
func TestController_Update_lockedFullDefinition(t *testing.T) {
	t.Parallel()
	h := newLockedHarness(t, "registries:"+corpRegistry+"packages:\n- name: example-org/tool@tool/v1.0.0\n  registry: corp\n", recordCorp(t))
	h.getter.versions["example-org/tool"] = "tool/v1.2.0"
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if got := h.read(t, "aqua.yaml"); !strings.Contains(got, "example-org/tool@tool/v1.2.0") {
		t.Errorf("aqua.yaml isn't updated:\n%s", got)
	}
	call := h.getter.calls["example-org/tool"]
	if call.source != versiongetter.SourceUpstream || call.pkg.VersionPrefix != "tool/" {
		t.Errorf("source %v, version_prefix %q, want SourceUpstream and tool/", call.source, call.pkg.VersionPrefix)
	}
	if !slices.Equal(h.installer.atCommit, []string{"corp@abc"}) || len(h.installer.atRef) != 0 {
		t.Errorf("read at a commit %v and at a ref %v, want corp@abc only", h.installer.atCommit, h.installer.atRef)
	}
}

// A registry the lock file has no record of is read at its ref rather than failing.
func TestController_Update_lockedUnrecordedRegistry(t *testing.T) {
	t.Parallel()
	h := newLockedHarness(t, "registries:"+corpRegistry+"packages:\n- name: example-org/tool@tool/v1.0.0\n  registry: corp\n", &lockfile.LockFile{})
	h.getter.versions["example-org/tool"] = "tool/v1.2.0"
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(h.installer.atRef, []string{"corp"}) {
		t.Errorf("read at a ref %v, want corp", h.installer.atRef)
	}
}

// A registry named standard that isn't aquaproj/aqua-registry is any other registry.
func TestController_Update_lockedForkNamedStandard(t *testing.T) {
	t.Parallel()
	h := newLockedHarness(t, `registries:
- name: standard
  type: github_content
  repo_owner: fork
  repo_name: aqua-registry
  ref: v1.0.0
  path: registry.yaml
packages:
- name: example-org/tool@tool/v1.0.0
`, &lockfile.LockFile{})
	h.installer.contents["standard"] = toolDefinition
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if call := h.getter.calls["example-org/tool"]; call.source != versiongetter.SourceUpstream || call.pkg.VersionPrefix != "tool/" {
		t.Errorf("source %v, version_prefix %q, want SourceUpstream and tool/", call.source, call.pkg.VersionPrefix)
	}
}

// A package that can't be looked at fails the run after the others are written, and the
// registries' refs aren't updated. A version upstream couldn't give isn't a failure.
func TestController_Update_lockedFailures(t *testing.T) {
	t.Parallel()
	h := newLockedHarness(t, "registries:\n- type: standard\n  ref: v4.0.0"+corpRegistry+`packages:
- name: cli/cli@v2.0.0
- name: other/g2-unreadable@v1.0.0
- name: other/missing-registry@v1.0.0
  registry: missing
- name: example-org/not-in-registry@v1.0.0
  registry: corp
- name: example-org/vars@v1.0.0
  registry: corp
- name: example-org/tool@tool/v1.0.0
  registry: corp
- version: v1.0.0
`, recordCorp(t))
	h.param.OnlyPackage = false
	h.getter.versions["cli/cli"] = "v2.1.0"
	h.getter.errs["other/g2-unreadable"] = true
	if err := h.run(t); err == nil {
		t.Fatal("an error must be returned")
	}
	if got := h.read(t, "aqua.yaml"); !strings.Contains(got, "cli/cli@v2.1.0") || !strings.Contains(got, "ref: v4.0.0") {
		t.Errorf("cli/cli must be updated and the refs left alone:\n%s", got)
	}
	if h.gh.latest != 0 {
		t.Error("the registries' latest releases must not be asked for")
	}
	for _, name := range []string{"other/missing-registry", "example-org/not-in-registry", "example-org/vars"} {
		if _, ok := h.getter.calls[name]; ok {
			t.Errorf("%s must fail before its versions are asked for", name)
		}
	}
	if _, ok := h.getter.calls["example-org/tool"]; !ok {
		t.Error("example-org/tool, whose version upstream didn't give, must still be asked for")
	}
}

// The packages are narrowed as before: update.enabled, the tags, and the choice with -i.
func TestController_Update_lockedFilters(t *testing.T) {
	t.Parallel()
	const cfg = `registries:
- type: standard
  ref: v4.0.0
packages:
- name: a/disabled@v1.0.0
  update:
    enabled: false
- name: a/excluded@v1.0.0
  tags: [skip]
- name: a/kept@v1.0.0
`
	h := newLockedHarness(t, cfg, &lockfile.LockFile{})
	h.param.ExcludedTags = map[string]struct{}{"skip": {}}
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if _, ok := h.getter.calls["a/kept"]; !ok || len(h.getter.calls) != 1 {
		t.Errorf("asked for %v, want a/kept only", h.getter.calls)
	}

	h = newLockedHarness(t, cfg, &lockfile.LockFile{})
	h.param.Insert = true
	h.idxs = []int{1}
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if _, ok := h.getter.calls["a/excluded"]; !ok || len(h.getter.calls) != 1 {
		t.Errorf("asked for %v, want a/excluded only", h.getter.calls)
	}
}

// A file imported by an imported file uses the registries of aqua.yaml and its lock file.
func TestController_Update_lockedNestedImport(t *testing.T) {
	t.Parallel()
	h := newLockedHarness(t, "registries:\n- type: standard\n  ref: v4.0.0"+corpRegistry+"packages:\n- import: a.yaml\n", recordCorp(t))
	writeFile(t, filepath.Join(h.dir, "a.yaml"), "packages:\n- import: b.yaml\n")
	writeFile(t, filepath.Join(h.dir, "b.yaml"), "packages:\n- name: cli/cli@v2.0.0\n- name: example-org/tool@tool/v1.0.0\n  registry: corp\n")
	h.getter.versions["cli/cli"] = "v2.1.0"
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if got := h.read(t, "b.yaml"); !strings.Contains(got, "cli/cli@v2.1.0") {
		t.Errorf("b.yaml isn't updated:\n%s", got)
	}
	if call := h.getter.calls["cli/cli"]; call.source != versiongetter.SourceG2 {
		t.Errorf("cli/cli source %v, want SourceG2", call.source)
	}
	if call := h.getter.calls["example-org/tool"]; call.source != versiongetter.SourceUpstream || call.pkg.VersionPrefix != "tool/" {
		t.Errorf("example-org/tool source %v, version_prefix %q", call.source, call.pkg.VersionPrefix)
	}
	if !slices.Equal(h.installer.atCommit, []string{"corp@abc"}) {
		t.Errorf("read at a commit %v, want corp@abc", h.installer.atCommit)
	}
}

// A broken lock file fails rather than being taken for none, except with --only-registry,
// which doesn't look at the packages.
func TestController_Update_lockedBrokenLockFile(t *testing.T) {
	t.Parallel()
	const cfg = "registries:\n- type: standard\n  ref: v4.0.0\npackages:\n- name: cli/cli@v2.0.0\n"
	h := newLockedHarness(t, cfg, "{")
	h.getter.versions["cli/cli"] = "v2.1.0"
	if err := h.run(t); err == nil {
		t.Fatal("an error must be returned")
	}
	if h.installer.installedAll != 0 || len(h.getter.calls) != 0 {
		t.Error("nothing must be read")
	}

	h = newLockedHarness(t, cfg, "{")
	h.param.OnlyPackage = false
	h.param.OnlyRegistry = true
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if got := h.read(t, "aqua.yaml"); !strings.Contains(got, "ref: v9.0.0") {
		t.Errorf("the registry's ref isn't updated:\n%s", got)
	}
}

// "aqua update <command>" looks for the version the same way: a package of another
// registry with its whole definition rather than the lock file's entry, and one of the
// standard registry from aqua-registry-g2 alone.
func TestController_Update_lockedCommand(t *testing.T) {
	t.Parallel()
	const cfg = "registries:\n- type: standard\n  ref: v4.0.0" + corpRegistry + `packages:
- name: example-org/tool@tool/v1.0.0
  registry: corp
- name: cli/cli@v2.0.0
`
	found := func(h *lockedHarness, name, rgst, version string) *which.FindResult {
		return &which.FindResult{
			ConfigFilePath: filepath.Join(h.dir, "aqua.yaml"),
			Package: &config.Package{
				Package: &aqua.Package{Name: name, Registry: rgst, Version: version},
				// What the lock file's entry gives: no version_prefix.
				PackageInfo: &registry.PackageInfo{Name: name, Type: "github_release"},
			},
		}
	}

	h := newLockedHarness(t, cfg, recordCorp(t))
	h.param.Args = []string{"tool"}
	h.which.FindResults = map[string]*which.FindResult{"tool": found(h, "example-org/tool", "corp", "tool/v1.0.0")}
	h.getter.versions["example-org/tool"] = "tool/v1.2.0"
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if got := h.read(t, "aqua.yaml"); !strings.Contains(got, "example-org/tool@tool/v1.2.0") {
		t.Errorf("aqua.yaml isn't updated:\n%s", got)
	}
	if call := h.getter.calls["example-org/tool"]; call.source != versiongetter.SourceUpstream || call.pkg.VersionPrefix != "tool/" {
		t.Errorf("source %v, version_prefix %q, want SourceUpstream and tool/", call.source, call.pkg.VersionPrefix)
	}

	// g2 has no version: nothing is updated and the run succeeds.
	h = newLockedHarness(t, cfg, recordCorp(t))
	h.param.Args = []string{"gh"}
	h.which.FindResults = map[string]*which.FindResult{"gh": found(h, "cli/cli", "standard", "v2.0.0")}
	if err := h.run(t); err != nil {
		t.Fatal(err)
	}
	if call := h.getter.calls["cli/cli"]; call.source != versiongetter.SourceG2 {
		t.Errorf("source %v, want SourceG2", call.source)
	}
	if got := h.read(t, "aqua.yaml"); !strings.Contains(got, "cli/cli@v2.0.0") {
		t.Errorf("aqua.yaml must be left alone:\n%s", got)
	}

	// g2 can't be read: the run fails.
	h = newLockedHarness(t, cfg, recordCorp(t))
	h.param.Args = []string{"gh"}
	h.which.FindResults = map[string]*which.FindResult{"gh": found(h, "cli/cli", "standard", "v2.0.0")}
	h.getter.errs["cli/cli"] = true
	if err := h.run(t); err == nil {
		t.Fatal("an error must be returned")
	}
}
