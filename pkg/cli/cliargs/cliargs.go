// Package cliargs provides common CLI argument structures and flag definitions
// shared across aqua CLI commands.
package cliargs

import (
	"github.com/urfave/cli/v3"
)

// GlobalArgs holds global CLI flags that are available to all commands.
type GlobalArgs struct {
	LogLevel                         string
	Config                           string
	DisableCosign                    bool
	DisableSLSA                      bool
	DisableGitHubArtifactAttestation bool
	DisableGitHubReleaseAttestation  bool
	Trace                            string
	CPUProfile                       string
}

// GlobalFlags returns the global CLI flags with destinations bound to the provided GlobalArgs.
func GlobalFlags(args *GlobalArgs) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Usage:       "log level",
			Sources:     cli.EnvVars("AQUA_LOG_LEVEL"),
			Destination: &args.LogLevel,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "configuration file path",
			Sources:     cli.EnvVars("AQUA_CONFIG"),
			Destination: &args.Config,
		},
		&cli.BoolFlag{
			Name:        "disable-cosign",
			Usage:       "Disable Cosign verification",
			Sources:     cli.EnvVars("AQUA_DISABLE_COSIGN"),
			Destination: &args.DisableCosign,
		},
		&cli.BoolFlag{
			Name:        "disable-slsa",
			Usage:       "Disable SLSA verification",
			Sources:     cli.EnvVars("AQUA_DISABLE_SLSA"),
			Destination: &args.DisableSLSA,
		},
		&cli.BoolFlag{
			Name:        "disable-github-artifact-attestation",
			Usage:       "Disable GitHub Artifact Attestations verification",
			Sources:     cli.EnvVars("AQUA_DISABLE_GITHUB_ARTIFACT_ATTESTATION"),
			Destination: &args.DisableGitHubArtifactAttestation,
		},
		&cli.BoolFlag{
			Name:        "disable-github-immutable-release",
			Usage:       "Disable GitHub Release Attestations verification",
			Sources:     cli.EnvVars("AQUA_DISABLE_GITHUB_IMMUTABLE_RELEASE"),
			Destination: &args.DisableGitHubReleaseAttestation,
		},
		&cli.StringFlag{
			Name:        "trace",
			Usage:       "trace output file path",
			Destination: &args.Trace,
		},
		&cli.StringFlag{
			Name:        "cpu-profile",
			Usage:       "cpu profile output file path",
			Destination: &args.CPUProfile,
		},
	}
}
