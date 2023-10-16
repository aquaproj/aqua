package fuzzyfinder

import (
	"fmt"
)

type Version struct {
	Name        string
	Version     string
	Description string
	URL         string
}

func PreviewVersion(v *Version) string {
	s := v.Version
	if v.Name != "" && v.Name != v.Version {
		s += fmt.Sprintf(" (%s)", v.Name)
	}
	if v.URL != "" || v.Description != "" {
		s += "\n"
	}
	if v.URL != "" {
		s += fmt.Sprintf("\n%s", v.URL)
	}
	if v.URL != "" {
		s += fmt.Sprintf("\n%s", v.Description)
	}
	return s
}

func (v *Version) Preview(w int) string {
	s := v.Version
	if v.Name != "" && v.Name != v.Version {
		s += fmt.Sprintf(" (%s)", v.Name)
	}
	if v.URL != "" || v.Description != "" {
		s += "\n"
	}
	if v.URL != "" {
		s += fmt.Sprintf("\n%s", v.URL)
	}
	if v.URL != "" {
		s += fmt.Sprintf("\n%s", formatPreview(v.Description, w/2-8)) //nolint:gomnd
	}
	return s
}

func (v *Version) Item() string {
	return v.Version
}

func ConvertStringsToItems(arr []string) []*Item {
	items := make([]*Item, len(arr))
	for i, a := range arr {
		items[i] = &Item{
			Item: a,
		}
	}
	return items
}
