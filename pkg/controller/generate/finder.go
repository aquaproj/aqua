package generate

import (
	"fmt"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
)

func (ctrl *Controller) launchFuzzyFinder(pkgs []*FindingPackage) ([]int, error) {
	return fuzzyfinder.FindMulti(pkgs, func(i int) string { //nolint:wrapcheck
		pkg := pkgs[i]
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
	},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 {
				return ""
			}
			pkg := pkgs[i]
			return fmt.Sprintf("%s\n\n%s\n%s",
				pkg.PackageInfo.GetName(),
				pkg.PackageInfo.GetLink(),
				formatDescription(pkg.PackageInfo.GetDescription(), w/2-8)) //nolint:gomnd
		}))
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
