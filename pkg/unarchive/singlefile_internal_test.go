package unarchive

import "testing"

func TestSingleFileName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		format string
		asset  string
		want   string
	}{
		{name: "gz by format", format: "gz", asset: "duckdb_cli-linux-amd64.gz", want: "duckdb_cli-linux-amd64"},
		{name: "gz by extension", asset: "foo_linux_amd64.gz", want: "foo_linux_amd64"},
		{name: "xz", format: "xz", asset: "foo-1.2.3.xz", want: "foo-1.2.3"},
		{name: "zstd", format: "zst", asset: "foo.zst", want: "foo"},
		// A tar carries entries with names of their own, whether or not it is
		// compressed, so nothing here can say what the package installs from.
		{name: "tar.gz by format", format: "tar.gz", asset: "foo.tar.gz"},
		{name: "tar.gz by extension", asset: "foo.tar.gz"},
		{name: "tgz", format: "tgz", asset: "foo.tgz"},
		{name: "zip", format: "zip", asset: "foo.zip"},
		{name: "no extension", asset: "foo_linux_amd64"},
		// The path is dropped: what is asked for is a name inside the extraction
		// directory.
		{name: "asset with a path", format: "gz", asset: "dist/foo.gz", want: "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := SingleFileName(tt.format, tt.asset); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// What says where a package installs from and what writes the file there have to agree,
// so they share the expression that names it.
func TestSingleFileName_agreesWithDecompress(t *testing.T) {
	t.Parallel()
	const asset = "duckdb_cli-linux-amd64.gz"
	if got, want := SingleFileName("gz", asset), decompressedName(asset); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
