package which

import (
	"context"
	"errors"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
)

func (c *MockController) Which(_ context.Context, _ *logrus.Entry, _ *config.Param, _ string) (*FindResult, error) {
	return c.FindResult, c.Err
}

type MockMultiController struct {
	FindResults map[string]*FindResult
}

func (c *MockMultiController) Which(_ context.Context, _ *logrus.Entry, _ *config.Param, exeName string) (*FindResult, error) {
	fr, ok := c.FindResults[exeName]
	if !ok {
		return nil, errors.New("command isn't found")
	}
	return fr, nil
}
