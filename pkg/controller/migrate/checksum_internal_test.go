package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// write puts a configuration where the test can point the migration at it.
func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aqua.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// What the settings said a lock file says, so they are what is taken out -- and
// supported_envs isn't, because it says which environments the lock file records.
func TestChecksumSettings(t *testing.T) {
	t.Parallel()
	path := writeConfig(t, `checksum:
  enabled: true
  require_checksum: true
  supported_envs:
    - darwin
packages:
  - name: cli/cli@v2.0.0
`)
	got, err := checksumSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"checksum.enabled: true", "checksum.require_checksum: true"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the settings are wrong (-want +got):\n%s", diff)
	}
}

// A configuration with nothing of the sort has nothing to take out, which is what a run
// after the first one finds.
func TestChecksumSettings_none(t *testing.T) {
	t.Parallel()
	path := writeConfig(t, "packages:\n  - name: cli/cli@v2.0.0\n")
	got, err := checksumSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("the settings are %v, want none", got)
	}
}

// The comment and the order survive, because aqua.yaml is a file somebody wrote.
func TestRemoveChecksumSettings(t *testing.T) {
	t.Parallel()
	path := writeConfig(t, `---
# the tools this repository needs
checksum:
  enabled: true
  require_checksum: true
  supported_envs:
    - darwin
registries:
  - type: standard
    ref: v4.567.0
packages:
  - name: cli/cli@v2.0.0
`)
	if err := removeChecksumSettings(path); err != nil {
		t.Fatal(err)
	}
	want := `---
# the tools this repository needs
checksum:
  supported_envs:
    - darwin
registries:
  - type: standard
    ref: v4.567.0
packages:
  - name: cli/cli@v2.0.0
`
	if diff := cmp.Diff(want, read(t, path)); diff != "" {
		t.Errorf("the configuration is wrong (-want +got):\n%s", diff)
	}
}

// A section that said nothing else goes with the settings rather than staying as a key with
// nothing under it.
func TestRemoveChecksumSettings_wholeSection(t *testing.T) {
	t.Parallel()
	path := writeConfig(t, `checksum:
  enabled: true
packages:
  - name: cli/cli@v2.0.0
`)
	if err := removeChecksumSettings(path); err != nil {
		t.Fatal(err)
	}
	want := "packages:\n  - name: cli/cli@v2.0.0\n"
	if diff := cmp.Diff(want, read(t, path)); diff != "" {
		t.Errorf("the configuration is wrong (-want +got):\n%s", diff)
	}
}

// One setting parses as a mapping value rather than a mapping, and says the same thing.
func TestRemoveChecksumSettings_oneSetting(t *testing.T) {
	t.Parallel()
	path := writeConfig(t, "checksum:\n  require_checksum: true\npackages: []\n")
	if err := removeChecksumSettings(path); err != nil {
		t.Fatal(err)
	}
	if got, want := read(t, path), "packages: []\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
