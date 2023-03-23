package installpackage

import (
	"context"
)

type Executor interface {
	GoInstall(ctx context.Context, path, gobin string) (int, error)
}
