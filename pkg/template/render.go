package template

import (
	"text/template"

	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type Artifact struct {
	Version         string
	SemVer          string
	OS              string
	Arch            string
	Format          string
	Asset           string
	AssetWithoutExt string
	Vars            map[string]any
}

func renderParam(art *Artifact, rt *runtime.Runtime) map[string]any {
	return map[string]any{
		"Version":         art.Version,
		"SemVer":          art.SemVer,
		"GOOS":            rt.GOOS,
		"GOARCH":          rt.GOARCH,
		"OS":              art.OS,
		"Arch":            art.Arch,
		"Format":          art.Format,
		"Asset":           art.Asset,
		"AssetWithoutExt": art.AssetWithoutExt,
		"Vars":            art.Vars,
	}
}

func Render(s string, art *Artifact, rt *runtime.Runtime) (string, error) {
	return Execute(s, renderParam(art, rt))
}

func RenderTemplate(tpl *template.Template, art *Artifact, rt *runtime.Runtime) (string, error) {
	return ExecuteTemplate(tpl, renderParam(art, rt))
}
