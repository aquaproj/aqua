package settoken

import "io"

type Controller struct {
	stdout io.Writer
}

func New(stdout io.Writer) *Controller {
	return &Controller{
		stdout: stdout,
	}
}
