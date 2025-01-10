package vacuum

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
)

type NilVacuumController struct{}

func (m *NilVacuumController) StorePackage(_ *logrus.Entry, _ *config.Package, _ string) error {
	return nil
}

func (m *NilVacuumController) Close(_ *logrus.Entry) error {
	return nil
}

func (m *NilVacuumController) Vacuum(_ *logrus.Entry) error {
	return nil
}
