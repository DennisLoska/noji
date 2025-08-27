package commands

import (
	"github.com/dennis/noji/internal/commands/output"
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
			output.Infof(output.ModeAuto, "config: %s\n", cfg)
			output.Infof(output.ModeAuto, "prompts: %s\n", prompts)
			return nil
		},
	}
}
