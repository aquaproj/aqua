package resolve

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
)

// renderSigning resolves the signing configuration for one environment.
//
// A registry writes the names of signature files as templates, the same way it
// writes an asset name: {{.OS}} and {{.Arch}} stand for what the environment turns
// out to be, after replacements have been applied to it. A lock file entry that kept
// the templates would need the replacements kept with it to expand them again, and
// an entry describing one environment has no business carrying the rule that picked
// it. Everything is rendered here instead, so the entry says what to download rather
// than how to work it out.
func renderSigning(pkg *config.Package, info *registry.PackageInfo, rt *runtime.Runtime, assetName string) (*signing, error) {
	art := pkg.TemplateArtifact(rt, assetName)
	out := &signing{}

	if info.Cosign.GetEnabled() {
		cos, err := renderCosign(info.Cosign, rt, art)
		if err != nil {
			return nil, err
		}
		out.cosign = cos
	}
	if info.SLSAProvenance.GetEnabled() {
		sp := *info.SLSAProvenance
		asset, err := renderPtr(sp.Asset, rt, art)
		if err != nil {
			return nil, fmt.Errorf("render the SLSA provenance asset: %w", err)
		}
		url, err := renderPtr(sp.URL, rt, art)
		if err != nil {
			return nil, fmt.Errorf("render the SLSA provenance URL: %w", err)
		}
		sp.Asset, sp.URL = asset, url
		out.slsa = &sp
	}
	if info.Minisign.GetEnabled() {
		m := *info.Minisign
		asset, err := renderPtr(m.Asset, rt, art)
		if err != nil {
			return nil, fmt.Errorf("render the minisign signature asset: %w", err)
		}
		url, err := renderPtr(m.URL, rt, art)
		if err != nil {
			return nil, fmt.Errorf("render the minisign signature URL: %w", err)
		}
		m.Asset, m.URL = asset, url
		out.minisign = &m
	}
	return out, nil
}

// signing is the rendered configuration of one environment.
type signing struct {
	cosign   *registry.Cosign
	slsa     *registry.SLSAProvenance
	minisign *registry.Minisign
}

func renderCosign(cos *registry.Cosign, rt *runtime.Runtime, art *template.Artifact) (*registry.Cosign, error) {
	out := *cos
	opts, err := cos.RenderOpts(rt, art)
	if err != nil {
		return nil, fmt.Errorf("render the cosign options: %w", err)
	}
	out.Opts = opts
	for _, f := range []struct {
		name string
		src  **registry.DownloadedFile
	}{
		{"signature", &out.Signature},
		{"certificate", &out.Certificate},
		{"key", &out.Key},
		{"bundle", &out.Bundle},
	} {
		rendered, err := renderDownloadedFile(*f.src, rt, art)
		if err != nil {
			return nil, fmt.Errorf("render the cosign %s: %w", f.name, err)
		}
		*f.src = rendered
	}
	return &out, nil
}

func renderDownloadedFile(f *registry.DownloadedFile, rt *runtime.Runtime, art *template.Artifact) (*registry.DownloadedFile, error) {
	if f == nil {
		return nil, nil //nolint:nilnil // nothing of this kind is configured
	}
	out := *f
	asset, err := renderPtr(f.Asset, rt, art)
	if err != nil {
		return nil, err
	}
	url, err := renderPtr(f.URL, rt, art)
	if err != nil {
		return nil, err
	}
	out.Asset, out.URL = asset, url
	return &out, nil
}

// renderPtr renders a template an optional field holds.
func renderPtr(s *string, rt *runtime.Runtime, art *template.Artifact) (*string, error) {
	if s == nil {
		return nil, nil //nolint:nilnil // the field isn't set
	}
	rendered, err := template.Render(*s, art, rt)
	if err != nil {
		return nil, fmt.Errorf("render %q: %w", *s, err)
	}
	return &rendered, nil
}
