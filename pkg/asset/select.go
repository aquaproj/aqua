package asset

import (
	"cmp"
	"fmt"
	"path"
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

// SelectOptions is what a definition says that helps tell a release's assets apart.
type SelectOptions struct {
	// Commands are the names of the package's executables.
	Commands []string
	// Format is the format the definition declares for an environment, or "" when it
	// declares none. It may be nil.
	Format func(goos, goarch, libc string) string
}

func (o *SelectOptions) commands() []string {
	if o == nil {
		return nil
	}
	return o.Commands
}

func (o *SelectOptions) format(goos, goarch, libc string) string {
	if o == nil || o.Format == nil {
		return ""
	}
	return o.Format(goos, goarch, libc)
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
// score beats a lower one, a binary beats a file named after it, an asset named for one of
// the package's commands beats the others, and of one build published in several formats
// the one in the format the definition declares wins, or else the format every machine
// opens. Anything left equal is a tie, which is reported
// rather than broken.
//
// The last segment of the package name is one of the commands too, which is how a release publishing several programs is read:
// kcp-dev/kcp's kcp_0.33.1_linux_amd64.tar.gz beside apigen_0.33.1_linux_amd64.tar.gz.
//
// A darwin build whose name says no architecture is for both, which is how a universal
// binary is named. An arm64 machine with no build of its own runs the amd64 one, on macOS
// by Rosetta 2 and on Windows by its emulation.
func Select(pkgName, version string, names []string, spellings Spellings, opts *SelectOptions) *Selection {
	o := &SelectOptions{Commands: append(slices.Clone(opts.commands()), path.Base(pkgName)), Format: opts.format}
	type key struct{ os, arch, libc string }
	groups := map[key][]*Selected{}
	sel := &Selection{}
	for _, name := range names {
		if !installable(pkgName, name, version) {
			continue
		}
		info := ParseAssetName(name, version, spellings)
		if info.OS == "" && hasWord(strings.ToLower(name), alpine) {
			// Alpine Linux is a distribution built on musl: snyk/cli's snyk-alpine-arm64 is
			// its musl build for linux/arm64.
			info.OS = osLinux
		}
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
		chosen, tied := Pick(candidates, o)
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
func SelectAs(goos string, goarchs []string, names []string, version string, spellings Spellings, opts *SelectOptions) *Selection {
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
		chosen, tied := Pick(candidates, opts)
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
func Pick(candidates []*Selected, opts *SelectOptions) (*Selected, []string) {
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
	named := slices.DeleteFunc(slices.Clone(best), func(s *Selected) bool { return !namedFor(s.Info, opts.commands()) })
	if len(named) == 1 {
		return named[0], nil
	}
	if format := inFormats(best, opts.format(best[0].OS, best[0].Arch, best[0].Libc)); format != nil {
		return format, nil
	}
	names := make([]string, len(best))
	for i, s := range best {
		names[i] = s.Name
	}
	slices.Sort(names)
	return nil, slices.Compact(names)
}

// formatPreference is the order archives of the same build are preferred in: the formats
// every machine can open first.
var formatPreference = []string{"tar.gz", "tgz", "zip", "tar.xz", "txz", "tar.zst", "tar.bz2", "tbz2", "tbz"} //nolint:gochecknoglobals

// inFormats returns, for candidates that are one build published in several formats, the
// one in the declared format, or else in the format preferred, or nil when they are more than
// that. koalaman/shellcheck publishes each build as a .tar.gz and a .tar.xz: they hold the
// same files, so either installs the same thing. tamasfe/taplo's Windows build is a .gz
// holding taplo-windows-x86_64 and a .zip holding taplo.exe, and the definition's files
// describe the .zip, which is what the format it declares says.
func inFormats(candidates []*Selected, declared string) *Selected {
	for _, c := range candidates[1:] {
		if c.Info.Template != candidates[0].Info.Template {
			return nil
		}
	}
	rank := func(s *Selected) int {
		if declared != "" && s.Info.Format == declared {
			return -1
		}
		if i := slices.Index(formatPreference, s.Info.Format); i >= 0 {
			return i
		}
		return len(formatPreference)
	}
	best := slices.MinFunc(candidates, func(a, b *Selected) int {
		return cmp.Or(cmp.Compare(rank(a), rank(b)), cmp.Compare(a.Info.Format, b.Info.Format))
	})
	return best
}

// namedFor says whether an asset is named for one of the commands: what its name says
// before the version or the platform is the command's name and nothing more.
// kcp_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}} is named for kcp, and
// kcp-front-proxy_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}} isn't.
func namedFor(info *AssetInfo, commands []string) bool {
	prefix, _, _ := strings.Cut(info.Template, "{")
	prefix = strings.TrimRight(prefix, "_-. ")
	if prefix == "" {
		return false
	}
	return slices.ContainsFunc(commands, func(c string) bool { return strings.EqualFold(prefix, c) })
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
// its own gets one. On Windows a build whose name says no architecture gets one too: it is
// the amd64 build or the 32-bit one, and Windows on arm64 runs either.
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
// template as well as a name: claude-{{.OS}}-{{.Arch}}-musl.{{.Format}} says musl, and so
// does a name for Alpine Linux, which is built on musl.
func LibcOf(goos, name string) string {
	if goos != osLinux {
		return ""
	}
	low := strings.ToLower(name)
	if strings.Contains(low, LibcMusl) || hasWord(low, alpine) {
		return LibcMusl
	}
	if strings.Contains(low, LibcGlibc) || hasWord(low, "gnu") {
		return LibcGlibc
	}
	return ""
}

// alpine is the word naming a build for Alpine Linux.
const alpine = "alpine"

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
