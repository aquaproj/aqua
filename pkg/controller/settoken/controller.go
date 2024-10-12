package settoken

import "io"

type Controller struct {
	stdin  io.Reader
	stdout io.Writer
}

func New(stdin io.Reader, stdout io.Writer) *Controller {
	return &Controller{
		stdin:  stdin,
		stdout: stdout,
	}
}
