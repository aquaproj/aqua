package g2_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/google/go-cmp/cmp"
)

func TestEncodePackageName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		pkg  string
		want string
	}{
		{name: "a slash", pkg: "cli/cli", want: "cli_2fcli"},
		{name: "two slashes", pkg: "ipinfo/cli/grepip", want: "ipinfo_2fcli_2fgrepip"},
		{
			// "~" can't appear in a git ref, so it has to be escaped rather than
			// left alone.
			name: "a character a ref can't hold",
			pkg:  "sr.ht/~charles/rq",
			want: "sr.ht_2f_7echarles_2frq",
		},
		{
			// The underscore is escaped too, which is what makes the mapping
			// reversible.
			name: "an underscore",
			pkg:  "sue445/plant_erd",
			want: "sue445_2fplant_5ferd",
		},
		{name: "nothing to escape", pkg: "aqua-registry.v2", want: "aqua-registry.v2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, g2.EncodePackageName(tt.pkg)); diff != "" {
				t.Errorf("EncodePackageName is wrong (-want +got):\n%s", diff)
			}
		})
	}
}

// TestBranchName_prefixPairs checks the pairs that turning a slash into two
// underscores could not represent: git refuses to hold refs/heads/a and
// refs/heads/a/b at once, and aqua-registry has 30 package pairs in that shape.
func TestBranchName_prefixPairs(t *testing.T) {
	t.Parallel()
	a := g2.BranchName("ipinfo/cli")
	b := g2.BranchName("ipinfo/cli/grepip")
	if a == b {
		t.Fatalf("the two packages encode to the same branch: %s", a)
	}
	for _, name := range []string{a, b} {
		for _, c := range name {
			if c == '/' {
				t.Errorf("a branch name must be a single ref segment, got %s", name)
			}
		}
	}
}

func TestPath(t *testing.T) {
	t.Parallel()
	if diff := cmp.Diff("versions/v2.1.0/registry-1.json", g2.Path("v2.1.0")); diff != "" {
		t.Errorf("Path is wrong (-want +got):\n%s", diff)
	}
}

// Encoding and decoding have to be exact inverses: the branch name is the only place
// a package's name is written on its branch.
func TestPackageName(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"cli/cli",
		"ipinfo/cli/grepip",
		"sr.ht/~charles/rq",
		"sue445/plant_erd",
		"aqua-registry.v2",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, ok := g2.PackageName(g2.BranchName(name))
			if !ok {
				t.Fatalf("%q didn't decode", name)
			}
			if diff := cmp.Diff(name, got); diff != "" {
				t.Errorf("the name is wrong (-want +got):\n%s", diff)
			}
		})
	}
}

// A branch that isn't a package's, or that no encoding could have produced, is
// reported rather than turned into a name.
func TestPackageName_notAPackage(t *testing.T) {
	t.Parallel()
	for _, branch := range []string{
		"main",
		"ar2_cli_2fcli",
		// A truncated escape.
		"pkg_cli_2",
		// Not hexadecimal.
		"pkg_cli_zzcli",
		// An escape of a character that would never have been escaped.
		"pkg_cli_61cli",
	} {
		t.Run(branch, func(t *testing.T) {
			t.Parallel()
			if got, ok := g2.PackageName(branch); ok {
				t.Errorf("%q decoded to %q, want a refusal", branch, got)
			}
		})
	}
}
