package vacuum

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
)

type MockVacuumController struct{}

func NewMockVacuumController() *MockVacuumController {
	return &MockVacuumController{}
}

func (m *MockVacuumController) StorePackage(logE *logrus.Entry, pkg *config.Package, pkgPath string) error {
	// Implementation of the mock method
	return nil
}

func (m *MockVacuumController) Close(logE *logrus.Entry) error {
	// Implementation of the mock method
	return nil
}

func (m *MockVacuumController) Vacuum(logE *logrus.Entry) error {
	// Implementation of the mock method
	return nil
}
