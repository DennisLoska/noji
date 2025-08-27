package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dennis/noji/internal/commands/output"
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
	cmd.AddCommand(newPREditCmd())
	cmd.AddCommand(newPRUpdateCmd())
	cmd.AddCommand(newReviewsPRCmd())
	cmd.AddCommand(newPRCommentsCmd())
	return cmd
}

func newPRCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a PR using opencode",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.Infof(output.ModeAuto, "Creating PR with model %s...\n", mustModel())
			defer output.Successf(output.ModeAuto, "Done.\n")
			return runPrompt("pr_create.txt")
		},
	}
}

func newPREditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit PR description for current branch using nvim",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.Infof(output.ModeAuto, "Editing PR description...\n")
			defer output.Successf(output.ModeAuto, "Done.\n")
			return runPREdit()
		},
	}
}

func newPRUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update a PR using opencode",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.Infof(output.ModeAuto, "Updating PR with model %s...\n", mustModel())
			defer output.Successf(output.ModeAuto, "Done.\n")
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

func mustModel() string {
	m, err := config.GetModel()
	if err != nil || m == "" {
		return "<unknown>"
	}
	return m
}

func runPREdit() error {
	// Get current branch
	branch, err := getCurrentBranch()
	if err != nil {
		return fmt.Errorf("get current branch: %w", err)
	}

	output.Infof(output.ModeAuto, "Current branch: %s\n", branch)

	// For now, create a temporary file with placeholder content
	// In a real implementation, this would fetch the actual PR description
	content := fmt.Sprintf(`# PR Description for branch: %s

## Summary
[Edit this PR description]

## Description
[Add your detailed description here]

## Next steps
[Add any next steps or TODOs]
`, branch)

	// Create temporary file
	tmpFile, err := createTempFile(content)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile) // Clean up

	output.Infof(output.ModeAuto, "Opening %s with nvim...\n", tmpFile)

	// Open with nvim
	if err := openWithNvim(tmpFile); err != nil {
		return fmt.Errorf("open with nvim: %w", err)
	}

	// Read the edited content
	editedContent, err := os.ReadFile(tmpFile)
	if err != nil {
		return fmt.Errorf("read edited file: %w", err)
	}

	output.Infof(output.ModeAuto, "Edited content:\n%s\n", string(editedContent))

	// TODO: In a real implementation, you would update the PR with the edited content
	output.Infof(output.ModeAuto, "Note: PR update functionality would be implemented here\n")

	return nil
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func createTempFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "noji-pr-edit-*.md")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func openWithNvim(filename string) error {
	cmd := exec.Command("nvim", filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
