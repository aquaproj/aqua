package generate

import (
	"fmt"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
)

type FuzzyFinder interface {
	Find(pkgs []*FindingPackage) ([]int, error)
}

type fuzzyFinder struct{}

func NewFuzzyFinder() FuzzyFinder {
	return &fuzzyFinder{}
}

func (finder *fuzzyFinder) Find(pkgs []*FindingPackage) ([]int, error) {
	return fuzzyfinder.FindMulti(pkgs, func(i int) string { //nolint:wrapcheck
		return find(pkgs[i])
	},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			return getPreview(pkgs[i], i, w)
		}))
}

func find(pkg *FindingPackage) string {
	files := pkg.PackageInfo.GetFiles()
	fileNames := make([]string, len(files))
	for i, file := range files {
		fileNames[i] = file.Name
	}
	fileNamesStr := strings.Join(fileNames, ", ")
	pkgName := pkg.PackageInfo.GetName()
	aliases := make([]string, len(pkg.PackageInfo.Aliases))
	for i, alias := range pkg.PackageInfo.Aliases {
		aliases[i] = alias.Name
	}
	item := pkgName
	if len(aliases) != 0 {
		item += " (" + strings.Join(aliases, ", ") + ")"
	}
	item += " (" + pkg.RegistryName + ")"
	if strings.HasSuffix(pkgName, "/"+fileNamesStr) || pkgName == fileNamesStr {
		return item
	}
	return item + " (" + fileNamesStr + ")"
}

func getPreview(pkg *FindingPackage, i, w int) string {
	if i < 0 {
		return ""
	}
	return fmt.Sprintf("%s\n\n%s\n%s",
		pkg.PackageInfo.GetName(),
		pkg.PackageInfo.GetLink(),
		formatDescription(pkg.PackageInfo.GetDescription(), w/2-8)) //nolint:gomnd
}

func formatDescription(desc string, w int) string {
	descRune := []rune(desc)
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
