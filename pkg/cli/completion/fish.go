package completion

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func (cm *command) fish(_ context.Context, _ *cli.Command) error {
	s, err := cm.cmd.ToFishCompletion()
	if err != nil {
		return fmt.Errorf("generate fish completion: %w", err)
	}
	fmt.Fprintln(cm.r.Stdout, s)
	return nil
}
