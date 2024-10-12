package settoken

import (
	"context"
	"fmt"
	"io"
	"strings"
	"syscall"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

func (c *Controller) Set(_ context.Context, logE *logrus.Entry, param *config.Param) error {
	text, err := c.get(logE, param)
	if err != nil {
		return fmt.Errorf("get a GitHub access Token: %w", err)
	}
	if err := github.SetTokenInKeyring(strings.TrimSpace(string(text))); err != nil {
		return fmt.Errorf("set a GitHub access Token to the secret store: %w", err)
	}
	return nil
}

func (c *Controller) get(logE *logrus.Entry, param *config.Param) ([]byte, error) {
	if param.Stdin {
		s, err := io.ReadAll(c.stdin)
		if err != nil {
			return nil, fmt.Errorf("read a GitHub access token from stdin: %w", err)
		}
		logE.Debug("read a GitHub access token from stdin")
		return s, nil
	}
	fmt.Fprint(c.stdout, "Enter a GitHub acccess token for aqua: ")
	text, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(c.stdout, "")
	if err != nil {
		return nil, fmt.Errorf("read a GitHub access Token from stdin: %w", err)
	}
	return text, nil
}
