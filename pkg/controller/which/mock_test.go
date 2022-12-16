package which_test

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/sirupsen/logrus"
)

type MockCosignVerifier struct {
	hasCosign bool
	err       error
}

func (mock *MockCosignVerifier) HasCosign() bool {
	return mock.hasCosign
}

func (mock *MockCosignVerifier) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error {
	return mock.err
}
