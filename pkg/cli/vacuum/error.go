package vacuum

import "errors"

var errVacuumTrackingDisabled = errors.New(`the vacuum command isn't available. Tracking is disabled`)
