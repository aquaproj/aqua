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

func TestPackageDir(t *testing.T) {
	t.Parallel()
	for id, want := range map[string]string{
		"1790772769": "pkgs/69/1790772769",
		"1790773000": "pkgs/00/1790773000",
		"7":          "pkgs/07/7",
	} {
		if got := g2.PackageDir(id); got != want {
			t.Errorf("PackageDir(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestPath(t *testing.T) {
	t.Parallel()
	if diff := cmp.Diff("versions/v2.1.0/registry-1.json", g2.Path("v2.1.0")); diff != "" {
		t.Errorf("Path is wrong (-want +got):\n%s", diff)
	}
}

// Encoding and decoding have to be exact inverses: the cache keeps a package's files under
// its encoded name.
func TestDecodePackageName(t *testing.T) {
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
			got, ok := g2.DecodePackageName(g2.EncodePackageName(name))
			if !ok {
				t.Fatalf("%q didn't decode", name)
			}
			if diff := cmp.Diff(name, got); diff != "" {
				t.Errorf("the name is wrong (-want +got):\n%s", diff)
			}
		})
	}
}

// What no encoding could have produced is reported rather than turned into a name.
func TestDecodePackageName_notEncoded(t *testing.T) {
	t.Parallel()
	for _, encoded := range []string{
		// A truncated escape.
		"cli_2",
		// Not hexadecimal.
		"cli_zzcli",
		// An escape of a character that would never have been escaped.
		"cli_61cli",
	} {
		t.Run(encoded, func(t *testing.T) {
			t.Parallel()
			if got, ok := g2.DecodePackageName(encoded); ok {
				t.Errorf("%q decoded to %q, want a refusal", encoded, got)
			}
		})
	}
}
