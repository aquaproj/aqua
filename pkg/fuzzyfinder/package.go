package fuzzyfinder

import (
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

type Package struct {
	PackageInfo  *registry.PackageInfo
	RegistryName string
	Version      string
}

func PreviewPackage(p *Package) string {
	return fmt.Sprintf("%s\n\n%s\n%s",
		p.PackageInfo.GetName(),
		p.PackageInfo.GetLink(),
		p.PackageInfo.Description)
}

func (p *Package) Preview(w int) string {
	return fmt.Sprintf("%s\n\n%s\n%s",
		p.PackageInfo.GetName(),
		p.PackageInfo.GetLink(),
		formatPreview(p.PackageInfo.Description, w/2-8)) //nolint:gomnd
}

func (p *Package) Item() string {
	files := p.PackageInfo.GetFiles()
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		if file.Name == "" {
			continue
		}
		fileNames = append(fileNames, file.Name)
	}
	fileNamesStr := strings.Join(fileNames, ", ")
	aliases := make([]string, 0, len(p.PackageInfo.Aliases))
	for _, alias := range p.PackageInfo.Aliases {
		if alias.Name == "" {
			continue
		}
		aliases = append(aliases, alias.Name)
	}
	pkgName := p.PackageInfo.GetName()
	item := pkgName
	if len(aliases) != 0 {
		item += " (" + strings.Join(aliases, ", ") + ")"
	}
	if p.RegistryName != registryStandard {
		item += " (" + p.RegistryName + ")"
	}
	if !strings.HasSuffix(pkgName, "/"+fileNamesStr) || pkgName == fileNamesStr {
		item += " [" + fileNamesStr + "]"
	}
	if len(p.PackageInfo.SearchWords) > 0 {
		item += ": " + strings.Join(p.PackageInfo.SearchWords, " ")
	}
	return item
}
