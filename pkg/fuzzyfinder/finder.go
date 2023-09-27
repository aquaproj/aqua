package fuzzyfinder

import (
	"fmt"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
)

var ErrAbort = fuzzyfinder.ErrAbort

type Item interface {
	Item() string
	Preview(w int) string
}

type Finder struct{}

func New() *Finder {
	return &Finder{}
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
