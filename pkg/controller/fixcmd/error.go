package fixcmd

import "errors"

// errOutOfDate is what --check reports with. A run that found something to do and was
// told not to do it has to say so in its exit status, or the job that runs it passes.
var errOutOfDate = errors.New("the configuration file is out of date")
