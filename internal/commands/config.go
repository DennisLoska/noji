package commands

import (
	"github.com/dennisloska/noji/internal/commands/output"
	"github.com/dennisloska/noji/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Manage configuration"}
	cmd.AddCommand(newConfigPathCmd())
	cmd.AddCommand(newConfigSetEditorCmd())
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

func newConfigSetEditorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-editor [editor]",
		Short: "Set the preferred editor in config (e.g. vim, vi, nvim, 'code -w')",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ed := args[0]
			if err := config.SetEditor(ed); err != nil {
				return err
			}
			output.Successf(output.ModeAuto, "Editor set to: %s\n", ed)
			return nil
		},
	}
}
