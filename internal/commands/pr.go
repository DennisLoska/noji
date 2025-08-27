package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dennis/noji/internal/config"
	"github.com/dennis/noji/internal/opencode"
	"github.com/spf13/cobra"
)

func newPRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Work with pull requests",
	}
	cmd.AddCommand(newPRCreateCmd())
	cmd.AddCommand(newPRUpdateCmd())
	cmd.AddCommand(newReviewsPRCmd())
	return cmd
}

func newPRCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a PR using opencode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrompt("pr_create.txt")
		},
	}
}

func newPRUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update a PR using opencode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrompt("pr_update.txt")
		},
	}
}

func runPrompt(promptFile string) error {
	model, err := config.GetModel()
	if err != nil {
		return err
	}
	promptsDir, err := config.PromptsDir()
	if err != nil {
		return err
	}
	p := filepath.Join(promptsDir, promptFile)
	b, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("read prompt file %s: %w", p, err)
	}
	return opencode.RunWithPrompt(model, string(b))
}
