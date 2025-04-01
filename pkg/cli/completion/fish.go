package completion

import (
	"fmt"

	"github.com/urfave/cli/v3"
)

func (cm *command) fish(c *cli.Context) error {
	s, err := c.App.ToFishCompletion()
	if err != nil {
		return fmt.Errorf("generate fish completion: %w", err)
	}
	fmt.Fprintln(cm.r.Stdout, s)
	return nil
}
