package fuzzyfinder

import (
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/ktr0731/go-fuzzyfinder"
)

var ErrAbort = fuzzyfinder.ErrAbort

type Item interface {
	Item() string
	Preview(w int) string
}

type Package struct {
	PackageInfo  *registry.PackageInfo
	RegistryName string
	Version      string
}

func (p *Package) Preview(w int) string {
	return fmt.Sprintf("%s\n\n%s\n%s",
		p.PackageInfo.GetName(),
		p.PackageInfo.GetLink(),
		formatDescription(p.PackageInfo.Description, w/2-8)) //nolint:gomnd
}

func (p *Package) Item() string {
	return find(p)
}

type Finder struct{}

func New() *Finder {
	return &Finder{}
}

func NewMock(idxs []int, err error) *MockFuzzyFinder {
	return &MockFuzzyFinder{
		idxs: idxs,
		err:  err,
	}
}

type MockFuzzyFinder struct {
	idxs []int
	err  error
}

func (f *MockFuzzyFinder) Find(items []Item, hasPreview bool) (int, error) {
	return f.idxs[0], f.err
}

func (f *MockFuzzyFinder) FindMulti(items []Item, hasPreview bool) ([]int, error) {
	return f.idxs, f.err
}

func (f *Finder) Find(items []Item, hasPreview bool) (int, error) {
	var opts []fuzzyfinder.Option
	if hasPreview {
		opts = []fuzzyfinder.Option{
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i < 0 {
					return "No item matches"
				}
				return items[i].Preview(w)
			}),
		}
	}
	return fuzzyfinder.Find(items, func(i int) string { //nolint:wrapcheck
		return items[i].Item()
	}, opts...)
}

func (f *Finder) FindMulti(items []Item, hasPreview bool) ([]int, error) {
	var opts []fuzzyfinder.Option
	if hasPreview {
		opts = []fuzzyfinder.Option{
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i < 0 {
					return "No item matches"
				}
				return items[i].Preview(w)
			}),
		}
	}
	return fuzzyfinder.FindMulti(items, func(i int) string { //nolint:wrapcheck
		return items[i].Item()
	}, opts...)
}

func find(pkg *Package) string {
	files := pkg.PackageInfo.GetFiles()
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		if file.Name == "" {
			continue
		}
		fileNames = append(fileNames, file.Name)
	}
	fileNamesStr := strings.Join(fileNames, ", ")
	aliases := make([]string, 0, len(pkg.PackageInfo.Aliases))
	for _, alias := range pkg.PackageInfo.Aliases {
		if alias.Name == "" {
			continue
		}
		aliases = append(aliases, alias.Name)
	}
	pkgName := pkg.PackageInfo.GetName()
	item := pkgName
	if len(aliases) != 0 {
		item += " (" + strings.Join(aliases, ", ") + ")"
	}
	if pkg.RegistryName != registryStandard {
		item += " (" + pkg.RegistryName + ")"
	}
	if !strings.HasSuffix(pkgName, "/"+fileNamesStr) || pkgName == fileNamesStr {
		item += " [" + fileNamesStr + "]"
	}
	if len(pkg.PackageInfo.SearchWords) > 0 {
		item += ": " + strings.Join(pkg.PackageInfo.SearchWords, " ")
	}
	return item
}

func getPreview(pkg *Package, i, w int) string {
	if i < 0 {
		return ""
	}
	return fmt.Sprintf("%s\n\n%s\n%s",
		pkg.PackageInfo.GetName(),
		pkg.PackageInfo.GetLink(),
		formatDescription(pkg.PackageInfo.Description, w/2-8)) //nolint:gomnd
}

func formatLine(line string, w int) string {
	descRune := []rune(line)
	lenDescRune := len(descRune)
	lineWidth := w - len([]rune("\n"))
	numOfLines := (lenDescRune / lineWidth) + 1
	descArr := make([]string, numOfLines)
	for i := 0; i < numOfLines; i++ {
		start := i * lineWidth
		end := start + lineWidth
		if i == numOfLines-1 {
			end = lenDescRune
		}
		descArr[i] = string(descRune[start:end])
	}
	return strings.Join(descArr, "\n")
}

func formatDescription(desc string, w int) string {
	lines := strings.Split(desc, "\n")
	arr := make([]string, len(lines))
	for i, line := range lines {
		arr[i] = formatLine(line, w)
	}
	return strings.Join(arr, "\n")
}
