package cp

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
)

type Installer interface {
	Install(ctx context.Context, logger *slog.Logger, param *config.Param) error
}
