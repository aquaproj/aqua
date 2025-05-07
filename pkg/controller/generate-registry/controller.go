package genrgst

import (
	"context"
	"io"

	"github.com/aquaproj/aqua/v2/pkg/cargo"
	"github.com/aquaproj/aqua/v2/pkg/controller/generate/output"
	"github.com/spf13/afero"
)

type Controller struct {
	stdout            io.Writer
	fs                afero.Fs
	github            RepositoriesService
	testdataOutputter TestdataOutputter
	cargoClient       CargoClient
}

type TestdataOutputter interface {
	Output(param *output.Param) error
}

func NewController(fs afero.Fs, gh RepositoriesService, testdataOutputter TestdataOutputter, cargoClient CargoClient, stdout io.Writer) *Controller {
	return &Controller{
		stdout:            stdout,
		fs:                fs,
		github:            gh,
		testdataOutputter: testdataOutputter,
		cargoClient:       cargoClient,
	}
}

type CargoClient interface {
	GetCrate(ctx context.Context, crate string) (*cargo.CratePayload, error)
	GetLatestVersion(ctx context.Context, crate string) (string, error)
}
