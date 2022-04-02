package config

import (
	"runtime"
	"strings"

	"github.com/aquaproj/aqua/pkg/template"
	"github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Package struct {
	Name     string `validate:"required" json:"name,omitempty"`
	Registry string `validate:"required" yaml:",omitempty" json:"registry,omitempty" jsonschema:"description=Registry name,example=foo,default=standard"`
	Version  string `validate:"required" yaml:",omitempty" json:"version,omitempty"`
	Import   string `yaml:",omitempty" json:"import,omitempty"`
}

func (pkg *Package) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias Package
	a := alias(*pkg)
	if err := unmarshal(&a); err != nil {
		return err
	}
	name, version := parseNameWithVersion(a.Name)
	if name != "" {
		a.Name = name
	}
	if version != "" {
		a.Version = version
	}
	*pkg = Package(a)
	if pkg.Registry == "" {
		pkg.Registry = RegistryTypeStandard
	}

	return nil
}

func parseNameWithVersion(name string) (string, string) {
	idx := strings.Index(name, "@")
	if idx == -1 {
		return name, ""
	}
	return name[:idx], name[idx+1:]
}

type Config struct {
	Packages   []*Package `validate:"dive" json:"packages"`
	Registries Registries `validate:"dive" json:"registries"`
}

type (
	PackageInfos []*PackageInfo
	Registries   map[string]*Registry
)

func (Registries) JSONSchema() *jsonschema.Schema {
	s := jsonschema.Reflect(&Registry{})
	return &jsonschema.Schema{
		Type:  "array",
		Items: s.Definitions["Registry"],
	}
}

const (
	PkgInfoTypeGitHubRelease = "github_release"
	PkgInfoTypeGitHubContent = "github_content"
	PkgInfoTypeGitHubArchive = "github_archive"
	PkgInfoTypeHTTP          = "http"
)

func (registries *Registries) UnmarshalYAML(unmarshal func(interface{}) error) error {
	arr := []*Registry{}
	if err := unmarshal(&arr); err != nil {
		return err
	}
	m := make(map[string]*Registry, len(arr))
	for _, registry := range arr {
		m[registry.Name] = registry
	}
	*registries = m
	return nil
}

func (pkgInfos *PackageInfos) ToMap() (map[string]*PackageInfo, error) {
	m := make(map[string]*PackageInfo, len(*pkgInfos))
	for _, pkgInfo := range *pkgInfos {
		pkgInfo := pkgInfo
		name := pkgInfo.GetName()
		if _, ok := m[name]; ok {
			return nil, logerr.WithFields(errPkgNameMustBeUniqueInRegistry, logrus.Fields{ //nolint:wrapcheck
				"package_name": name,
			})
		}
		m[name] = pkgInfo
	}
	return m, nil
}

type FormatOverride struct {
	GOOS   string `json:"goos" jsonschema:"example=darwin,example=linux"`
	Format string `yaml:"format" json:"format" jsonschema:"example=tar.gz,example=raw"`
}

type File struct {
	Name string             `validate:"required" json:"name,omitempty"`
	Src  *template.Template `json:"src,omitempty"`
}

func getArch(rosetta2 bool, replacements map[string]string) string {
	if rosetta2 && runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		// Rosetta 2
		return replace("amd64", replacements)
	}
	return replace(runtime.GOARCH, replacements)
}

func (file *File) RenderSrc(pkg *Package, pkgInfo *PackageInfo) (string, error) {
	return file.Src.Execute(map[string]interface{}{ //nolint:wrapcheck
		"Version":  pkg.Version,
		"GOOS":     runtime.GOOS,
		"GOARCH":   runtime.GOARCH,
		"OS":       replace(runtime.GOOS, pkgInfo.GetReplacements()),
		"Arch":     getArch(pkgInfo.GetRosetta2(), pkgInfo.GetReplacements()),
		"Format":   pkgInfo.GetFormat(),
		"FileName": file.Name,
	})
}

type Param struct {
	ConfigFilePath string
	LogLevel       string
	OnlyLink       bool
	IsTest         bool
	All            bool
	Insert         bool
	File           string
	GlobalConfigs  []string
	AQUAVersion    string
}
