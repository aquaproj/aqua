package which

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
)

func (c *MockController) Which(ctx context.Context, logger *slog.Logger, param *config.Param, exeName string) (*FindResult, error) {
	return c.FindResult, c.Err
}

type MockMultiController struct {
	FindResults map[string]*FindResult
}

func (c *MockMultiController) Which(ctx context.Context, logger *slog.Logger, param *config.Param, exeName string) (*FindResult, error) {
	fr, ok := c.FindResults[exeName]
	if !ok {
		return nil, errors.New("command isn't found")
	}
	return fr, nil
}
