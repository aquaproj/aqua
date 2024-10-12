package settoken

import (
	"context"
	"fmt"
	"strings"
	"syscall"

	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

func (c *Controller) Set(ctx context.Context, logE *logrus.Entry) error {
	fmt.Fprint(c.stdout, "Enter a GitHub acccess token for aqua: ")
	text, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("read a GitHub Access Token from stdin: %w", err)
	}
	fmt.Fprintln(c.stdout, "")
	if err := github.SetTokenInKeyring(strings.TrimSpace(string(text))); err != nil {
		return fmt.Errorf("set a GitHub Access Token to the secret store: %w", err)
	}
	return nil
}
