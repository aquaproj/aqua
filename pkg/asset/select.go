package asset

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
)

// Libc values an asset name can say.
const (
	LibcGlibc = "glibc"
	LibcMusl  = "musl"
)

// Selected is the asset a release publishes for one environment.
type Selected struct {
	// Name is the asset's name as the release lists it.
	Name string
	// Info is how the name was read. Its Replacements are how the name spells the OS
	// and the architecture, which is what a template written for the package, such as
	// a signature's name, has to be rendered with to name the release's file.
	Info *AssetInfo
	OS   string
	Arch string
	// Libc is the libc the name says the build is for, or "" when it says none. It is
	// only ever set on Linux.
	Libc string
	// Emulated says the asset is the amd64 build, chosen for an arm64 environment that
	// runs it: Rosetta 2 on macOS, Windows' emulation on Windows.
	Emulated bool
}

// Env is the environment the asset is for, as os/arch.
func (s *Selected) Env() string {
	return s.OS + "/" + s.Arch
}

// Tie is a set of assets that nothing in their names tells apart for one environment.
//
// A release can publish two builds for the same platform that differ in something the
// parser doesn't read, such as one for particular hardware beside the ordinary one
// (ollama's -rocm) or one for older CPUs (bun's -baseline). Picking one would be a guess,
// so the choice is left to whoever writes the package's filter.
type Tie struct {
	OS    string
	Arch  string
	Libc  string
	Names []string
}

func (t *Tie) String() string {
	env := t.OS + "/" + t.Arch
	if t.Libc != "" {
		env += " (" + t.Libc + ")"
	}
	return fmt.Sprintf("%s: %s", env, strings.Join(t.Names, ", "))
}

// Selection is what a release publishes, by environment.
type Selection struct {
	// Assets has one entry per os, arch and libc, sorted.
	Assets []*Selected
	// Ties are the environments whose assets nothing tells apart. They have no entry in
	// Assets.
	Ties []*Tie
	// Unnamed are the assets whose name says no OS, sorted. They are the release's, but
	// for no environment the name says.
	Unnamed []string
}

// Select picks the asset of each environment from a release's asset names.
//
// Everything is read off the names: the OS, the architecture, the libc and the format.
// Assets for the same os, arch and libc are ranked, and a release publishing a glibc and
// a musl build for one platform has both selected, as builds for different machines
// rather than rivals. The ranking is this: an archive beats a bare executable, a higher
// score beats a lower one, and a binary beats a file named after it. Anything left equal
// is a tie, which is reported rather than broken.
//
// A darwin build whose name says no architecture is for both, which is how a universal
// binary is named. An arm64 machine with no build of its own runs the amd64 one, on macOS
// always and on Windows when the name says amd64.
func Select(pkgName, version string, names []string, spellings Spellings) *Selection {
	type key struct{ os, arch, libc string }
	groups := map[key][]*Selected{}
	sel := &Selection{}
	for _, name := range names {
		if !installable(pkgName, name, version) {
			continue
		}
		info := ParseAssetName(name, version, spellings)
		if info.OS == "" {
			sel.Unnamed = append(sel.Unnamed, name)
			continue
		}
		libc := LibcOf(info.OS, name)
		for _, arch := range archsOf(info) {
			k := key{info.OS, arch, libc}
			groups[k] = append(groups[k], &Selected{Name: name, Info: info, OS: info.OS, Arch: arch, Libc: libc})
		}
	}
	for k, candidates := range groups {
		chosen, tied := Pick(candidates)
		if tied != nil {
			sel.Ties = append(sel.Ties, &Tie{OS: k.os, Arch: k.arch, Libc: k.libc, Names: tied})
			continue
		}
		sel.Assets = append(sel.Assets, chosen)
	}
	sel.Assets = append(sel.Assets, emulated(sel.Assets, sel.Ties)...)
	slices.SortFunc(sel.Assets, func(a, b *Selected) int {
		return cmp.Or(cmp.Compare(a.OS, b.OS), cmp.Compare(a.Arch, b.Arch), cmp.Compare(a.Libc, b.Libc))
	})
	slices.SortFunc(sel.Ties, func(a, b *Tie) int {
		return cmp.Or(cmp.Compare(a.OS, b.OS), cmp.Compare(a.Arch, b.Arch), cmp.Compare(a.Libc, b.Libc))
	})
	slices.Sort(sel.Unnamed)
	return sel
}

// SelectAs picks the asset of each of the given environments among assets whose name says
// no OS, which something other than the name says are for that OS.
//
// It is how a definition places such an asset: koalaman/shellcheck's Windows build is
// shellcheck-v0.11.0.zip. Each name is read as the OS's, and one saying no architecture is
// for every one of the environments.
func SelectAs(goos string, goarchs []string, names []string, version string, spellings Spellings) *Selection {
	sel := &Selection{}
	for _, goarch := range goarchs {
		var candidates []*Selected
		for _, name := range names {
			info := ParseAssetName(name, version, spellings)
			info.OS = goos
			if !info.NoArch && info.Arch != goarch {
				continue
			}
			candidates = append(candidates, &Selected{Name: name, Info: info, OS: goos, Arch: goarch, Libc: LibcOf(goos, name)})
		}
		if len(candidates) == 0 {
			continue
		}
		chosen, tied := Pick(candidates)
		if tied != nil {
			sel.Ties = append(sel.Ties, &Tie{OS: goos, Arch: goarch, Names: tied})
			continue
		}
		sel.Assets = append(sel.Assets, chosen)
	}
	return sel
}

// Pick returns the best of the candidates for one environment, or, when the best can't be
// told apart, nil and their names.
func Pick(candidates []*Selected) (*Selected, []string) {
	best := candidates
	if archives := slices.DeleteFunc(slices.Clone(best), isRaw); len(archives) != 0 {
		best = archives
	}
	top := slices.MaxFunc(best, func(a, b *Selected) int { return cmp.Compare(a.Info.Score, b.Info.Score) }).Info.Score
	best = slices.DeleteFunc(slices.Clone(best), func(s *Selected) bool { return s.Info.Score != top })
	best = slices.DeleteFunc(slices.Clone(best), func(s *Selected) bool {
		return slices.ContainsFunc(best, func(of *Selected) bool { return of != s && namedAfter(s.Info, of.Info) })
	})
	if len(best) == 1 {
		return best[0], nil
	}
	names := make([]string, len(best))
	for i, s := range best {
		names[i] = s.Name
	}
	slices.Sort(names)
	return nil, slices.Compact(names)
}

func isRaw(s *Selected) bool {
	return s.Info.Format == "" || s.Info.Format == formatRaw
}

// archsOf is the architectures an asset is for. A name saying none is the amd64 build,
// except on macOS, where a build named for no architecture is a universal one.
func archsOf(info *AssetInfo) []string {
	if info.NoArch && info.OS == osDarwin {
		return []string{archAmd64, archArm64}
	}
	return []string{info.Arch}
}

// emulated returns the arm64 environments that run the amd64 build, for the OSes that
// can: macOS by Rosetta 2, Windows by its emulation. Only an environment with no build of
// its own gets one, and on Windows only a build whose name says amd64, since a name
// saying nothing can as well be the 32-bit build.
func emulated(assets []*Selected, ties []*Tie) []*Selected {
	have := map[string]struct{}{}
	for _, a := range assets {
		have[a.Env()] = struct{}{}
	}
	for _, t := range ties {
		have[t.OS+"/"+t.Arch] = struct{}{}
	}
	var out []*Selected
	for _, a := range assets {
		if a.Arch != archAmd64 || (a.OS != osDarwin && a.OS != osWindows) {
			continue
		}
		if a.OS == osWindows && a.Info.NoArch {
			continue
		}
		if _, ok := have[a.OS+"/"+archArm64]; ok {
			continue
		}
		e := *a
		e.Arch = archArm64
		e.Emulated = true
		out = append(out, &e)
	}
	return out
}

// LibcOf is the libc a Linux asset's name says it is built against, or "". It reads a
// template as well as a name: claude-{{.OS}}-{{.Arch}}-musl.{{.Format}} says musl.
func LibcOf(goos, name string) string {
	if goos != osLinux {
		return ""
	}
	low := strings.ToLower(name)
	if strings.Contains(low, LibcMusl) {
		return LibcMusl
	}
	if strings.Contains(low, LibcGlibc) || hasWord(low, "gnu") {
		return LibcGlibc
	}
	return ""
}

// hasWord says whether a lower-cased name has the word, delimited by anything but a
// letter or a digit.
func hasWord(low, word string) bool {
	words := strings.FieldsFunc(low, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	return slices.Contains(words, word)
}

// installable says whether an asset can be what a package installs: not a build Exclude
// rules out, a checksum file, or a signature, certificate or attestation published beside
// a build.
func installable(pkgName, name, version string) bool {
	if Exclude(pkgName, name) {
		return false
	}
	low := strings.ToLower(name)
	for _, ext := range []string{
		".sig", ".asc", ".pem", ".cert", ".crt", ".bundle", ".sigstore", ".sigstore.json",
		".minisig", ".gpg", ".gpgsig", ".intoto.jsonl", ".pub", ".sha256", ".sha512", ".sha1", ".md5",
	} {
		if strings.HasSuffix(low, ext) {
			return false
		}
	}
	return checksum.GetChecksumConfigFromFilename(name, version) == nil
}
