package genrgst

import (
	"context"
	"io"

	"github.com/aquaproj/aqua/v2/pkg/cargo"
	"github.com/aquaproj/aqua/v2/pkg/controller/generate/output"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge"
)

type Controller struct {
	stdout            io.Writer
	github            RepositoriesService
	testdataOutputter TestdataOutputter
	cargoClient       CargoClient
	gitlab            GitLabClient
}

type TestdataOutputter interface {
	Output(param *output.Param) error
}

func NewController(gh RepositoriesService, testdataOutputter TestdataOutputter, cargoClient CargoClient, gitlabClient GitLabClient, stdout io.Writer) *Controller {
	return &Controller{
		stdout:            stdout,
		github:            gh,
		testdataOutputter: testdataOutputter,
		cargoClient:       cargoClient,
		gitlab:            gitlabClient,
	}
}

// GitLabClient reads a project on a GitLab instance: what it says it is, and one of its
// releases.
type GitLabClient interface {
	GetDescription(ctx context.Context, host, project string) (string, error)
	GetRelease(ctx context.Context, host, project, tag string) (*forge.Release, error)
}

type CargoClient interface {
	GetCrate(ctx context.Context, crate string) (*cargo.CratePayload, error)
	GetLatestVersion(ctx context.Context, crate string) (string, error)
}
