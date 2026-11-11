package checksum_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
)

func TestChecksum_HasVerified(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		chk   *checksum.Checksum
		kinds []string
		exp   bool
	}{
		{
			name:  "no entry",
			kinds: []string{"cosign"},
		},
		{
			name:  "no record",
			chk:   &checksum.Checksum{},
			kinds: []string{"cosign"},
		},
		{
			name:  "all recorded",
			chk:   &checksum.Checksum{Verified: []string{"cosign", "github_artifact_attestations"}},
			kinds: []string{"github_artifact_attestations", "cosign"},
			exp:   true,
		},
		{
			name:  "a kind missing",
			chk:   &checksum.Checksum{Verified: []string{"cosign"}},
			kinds: []string{"cosign", "slsa_provenance"},
		},
		{
			name: "nothing asked",
			chk:  &checksum.Checksum{},
			exp:  true,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if got := d.chk.HasVerified(d.kinds); got != d.exp {
				t.Fatalf("HasVerified = %v, wanted %v", got, d.exp)
			}
		})
	}
}

func TestChecksums_verified(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "aqua-checksums.json")
	if err := os.WriteFile(p, []byte(`{
  "checksums": [
    {
      "id": "github_release/github.com/aquaproj/ar2/v0.15.14/ar2_linux_amd64.tar.gz",
      "checksum": "AAAA",
      "algorithm": "sha256",
      "verified": ["github_artifact_attestations"]
    },
    {
      "id": "github_release/github.com/cli/cli/v2.17.0/gh_2.17.0_linux_amd64.tar.gz",
      "checksum": "BBBB",
      "algorithm": "sha256"
    }
  ]
}
`), 0o600); err != nil {
		t.Fatal(err)
	}
	checksums := checksum.New()
	if err := checksums.ReadFile(p); err != nil {
		t.Fatal(err)
	}
	if chk := checksums.Get("github_release/github.com/aquaproj/ar2/v0.15.14/ar2_linux_amd64.tar.gz"); !chk.HasVerified([]string{"github_artifact_attestations"}) {
		t.Fatalf("the record is lost: %+v", chk)
	}
	if chk := checksums.Get("github_release/github.com/cli/cli/v2.17.0/gh_2.17.0_linux_amd64.tar.gz"); chk.Verified != nil {
		t.Fatalf("an entry without the record must stay so: %+v", chk)
	}

	checksums.Set("github_release/github.com/suzuki-shunsuke/pinact/v3.0.0/pinact_linux_amd64.tar.gz", &checksum.Checksum{
		ID:        "github_release/github.com/suzuki-shunsuke/pinact/v3.0.0/pinact_linux_amd64.tar.gz",
		Checksum:  "CCCC",
		Algorithm: "sha256",
	})
	if err := checksums.UpdateFile(p); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	// Only the entry that recorded the verification writes the field.
	if n := strings.Count(string(b), `"verified"`); n != 1 {
		t.Fatalf(`"verified" is written %d times:\n%s`, n, b)
	}
}
