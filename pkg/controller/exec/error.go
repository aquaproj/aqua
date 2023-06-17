package exec

import "errors"

var errExecNotFoundDisableLazyInstall = errors.New(`the executable file isn't installed yet. Lazy Install is disabled`)
