package installpackage

import (
	"context"
)

type Executor interface {
	GoInstall(ctx context.Context, path, gobin string) (int, error)
	HdiutilAttach(ctx context.Context, dmgPath, mountPoint string) (int, error)
	HdiutilDetach(ctx context.Context, mountPath string) (int, error)
}
