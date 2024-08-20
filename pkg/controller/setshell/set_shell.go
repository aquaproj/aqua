package setshell

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
)

//go:embed zsh.sh
var zshScript []byte

//go:embed bash.sh
var bashScript []byte

func (c *Controller) SetShell(ctx context.Context, logE *logrus.Entry, param *config.Param, shellType string) error {
	switch shellType {
	case "":
		return errors.New("the argument shell type is required")
	case "bash":
		fmt.Fprintf(c.stdout, string(bashScript))
	case "zsh":
		fmt.Fprintf(c.stdout, string(zshScript))
	default:
		return errors.New(`supported shell types are 'bash' and 'zsh'`)
	}
	return nil
}
