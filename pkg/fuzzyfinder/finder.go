package fuzzyfinder

import (
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
)

var ErrAbort = fuzzyfinder.ErrAbort

type Item struct {
	Item    string
	Preview string
}

type Finder struct{}

func New() *Finder {
	return &Finder{}
}

func (f *Finder) Find(items []*Item, hasPreview bool) (int, error) {
	var opts []fuzzyfinder.Option
	if hasPreview {
		opts = []fuzzyfinder.Option{
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i < 0 {
					return "No item matches"
				}
				return formatPreview(items[i].Preview, w)
			}),
		}
	}
	return fuzzyfinder.Find(items, func(i int) string { //nolint:wrapcheck
		return items[i].Item
	}, opts...)
}

func (f *Finder) FindMulti(items []*Item, hasPreview bool) ([]int, error) {
	var opts []fuzzyfinder.Option
	if hasPreview {
		opts = []fuzzyfinder.Option{
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i < 0 {
					return "No item matches"
				}
				return formatPreview(items[i].Preview, w)
			}),
		}
	}
	return fuzzyfinder.FindMulti(items, func(i int) string { //nolint:wrapcheck
		return items[i].Item
	}, opts...)
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

func formatPreview(desc string, w int) string {
	lines := strings.Split(desc, "\n")
	arr := make([]string, len(lines))
	for i, line := range lines {
		arr[i] = formatLine(line, w)
	}
	return strings.Join(arr, "\n")
}
