package settoken

import (
	"context"
	"fmt"
	"syscall"

	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

func (c *Controller) Set(ctx context.Context, logE *logrus.Entry) error {
	fmt.Print("Enter a GitHub acccess token for aqua: ")
	text, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("read a GitHub Access Token from stdin: %w", err)
	}
	fmt.Println("")
	github.SetTokenInKeyring(string(text))
	return nil
}
