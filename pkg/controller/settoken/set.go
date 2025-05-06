package settoken

import (
	"fmt"
	"io"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

func (c *Controller) Set(logE *logrus.Entry) error {
	text, err := c.get(logE)
	if err != nil {
		return fmt.Errorf("get a GitHub access Token: %w", err)
	}
	if err := c.tokenManager.Set(strings.TrimSpace(string(text))); err != nil {
		return fmt.Errorf("set a GitHub access Token to the secret store: %w", err)
	}
	return nil
}

func (c *Controller) get(logE *logrus.Entry) ([]byte, error) {
	if c.param.IsStdin {
		s, err := io.ReadAll(c.param.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read a GitHub access token from stdin: %w", err)
		}
		logE.Debug("read a GitHub access token from stdin")
		return s, nil
	}
	text, err := c.term.ReadPassword()
	if err != nil {
		return nil, fmt.Errorf("read a GitHub access Token from terminal: %w", err)
	}
	return text, nil
}

type PasswordReader struct {
	stdout io.Writer
}

func NewPasswordReader(stdout io.Writer) *PasswordReader {
	return &PasswordReader{
		stdout: stdout,
	}
}

func (p *PasswordReader) ReadPassword() ([]byte, error) {
	fmt.Fprint(p.stdout, "Enter a GitHub access token: ")
	// The conversion is necessary for Windows
	// https://pkg.go.dev/syscall?GOOS=windows#pkg-variables
	b, err := term.ReadPassword(int(syscall.Stdin)) //nolint:unconvert
	fmt.Fprintln(p.stdout, "")
	if err != nil {
		return nil, fmt.Errorf("read a GitHub access token from terminal: %w", err)
	}
	return b, nil
}
