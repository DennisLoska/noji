package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dennis/noji/internal/opencode"
)

func newModelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "List available models",
		RunE: func(cmd *cobra.Command, args []string) error {
			models, err := opencode.ListModels()
			if err != nil {
				return err
			}
			for _, m := range models {
				fmt.Fprintln(cmd.OutOrStdout(), m)
			}
			return nil
		},
	}
	return cmd
}
