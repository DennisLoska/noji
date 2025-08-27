package commands

import (
	"errors"

	"github.com/dennisloska/noji/internal/commands/output"
	"github.com/dennisloska/noji/internal/config"
	"github.com/dennisloska/noji/internal/opencode"
	"github.com/spf13/cobra"
)

func newUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <model>",
		Short: "Select a model to use",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			model := args[0]
			models, err := opencode.ListModels()
			if err != nil {
				return err
			}
			if !contains(models, model) {
				return errors.New("model not found in available models")
			}
			if err := config.SetModel(model); err != nil {
				return err
			}
			output.Successf(output.ModeAuto, "Selected model: %s\n", model)
			return nil
		},
	}
	return cmd
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
