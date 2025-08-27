package commands

import (
	"fmt"

	"github.com/dennis/noji/internal/config"
	"github.com/spf13/cobra"
)

func newCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the currently selected model",
		RunE: func(cmd *cobra.Command, args []string) error {
			model, err := config.GetModel()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), model)
			return nil
		},
	}
}
