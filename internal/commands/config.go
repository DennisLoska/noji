package commands

import (
	"fmt"

	"github.com/dennis/noji/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Manage configuration"}
	cmd.AddCommand(newConfigPathCmd())
	return cmd
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print config and prompts paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, prompts, err := config.EnsureConfig()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "config:", cfg)
			fmt.Fprintln(cmd.OutOrStdout(), "prompts:", prompts)
			return nil
		},
	}
}
