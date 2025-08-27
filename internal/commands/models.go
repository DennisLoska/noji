package commands

import (
	"github.com/dennisloska/noji/internal/commands/output"
	"github.com/dennisloska/noji/internal/config"
	"github.com/dennisloska/noji/internal/opencode"
	"github.com/spf13/cobra"
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
			current, _ := config.GetModel()
			for _, m := range models {
				if m == current && m != "" {
					output.Successf(output.ModeAuto, "%s\n", m)
				} else {
					output.Printf(output.ModeAuto, "%s\n", m)
				}
			}
			return nil
		},
	}
	return cmd
}
