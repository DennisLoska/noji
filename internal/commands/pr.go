package commands

import (
	"bytes"
	"encoding/json"
	"errors"
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

type ghViewPR struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func newPRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Work with pull requests",
	}
	cmd.AddCommand(newPRCreateCmd())
	cmd.AddCommand(newPREditCmd())
	cmd.AddCommand(newPRUpdateCmd())
	cmd.AddCommand(newPRCommentsCmd())
	cmd.AddCommand(newReviewsPRCmd())
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
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit PR fields for current branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.Infof(output.ModeAuto, "Editing PR description...\n")
			defer output.Successf(output.ModeAuto, "Done.\n")
			return runPREdit()
		},
	}
	cmd.AddCommand(newPREditTitleSubCmd())
	cmd.AddCommand(newPREditBodySubCmd())
	return cmd
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

func newPREditTitleSubCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "title",
		Short: "Edit PR title for current branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.Infof(output.ModeAuto, "Editing PR title...\n")
			defer output.Successf(output.ModeAuto, "Done.\n")
			return runPREditTitle()
		},
	}
}

func newPREditBodySubCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "body",
		Short: "Edit PR body for current branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.Infof(output.ModeAuto, "Editing PR body...\n")
			defer output.Successf(output.ModeAuto, "Done.\n")
			return runPREditBody()
		},
	}
}

func ensureGh() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return errors.New("GitHub CLI 'gh' not found in PATH")
	}
	return nil
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

func getPRForCurrentBranch(branch string) (*ghViewPR, error) {
	// Use gh to view PR for current branch
	cmd := exec.Command("gh", "pr", "view", "--json", "number,title,body")
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		// If gh suggests there is no PR, return nil without error
		msg := errb.String()
		if strings.Contains(msg, "no open pull requests") || strings.Contains(msg, "no pull requests found") {
			return nil, nil
		}
		return nil, fmt.Errorf("gh pr view failed: %v: %s", err, msg)
	}
	var pr ghViewPR
	if err := json.Unmarshal(out.Bytes(), &pr); err != nil {
		return nil, fmt.Errorf("parse gh pr view json: %w", err)
	}
	return &pr, nil
}

func formatPRForEdit(body string) string {
	// Write the body as-is; no headers or comments.
	return body
}

// parseEditedPR removed: buffer is treated as opaque body only

func updatePRBody(number int, body string) error {
	args := []string{"pr", "edit", fmt.Sprintf("%d", number)}
	// Use --body-file to preserve formatting exactly
	tmp, err := createTempFile(body)
	if err != nil {
		return fmt.Errorf("create temp body file: %w", err)
	}
	defer os.Remove(tmp)
	args = append(args, "--body-file", tmp)
	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh pr edit failed: %w", err)
	}
	return nil
}

func mustModel() string {
	m, err := config.GetModel()
	if err != nil || m == "" {
		return "<unknown>"
	}
	return m
}

func runPREdit() error {
	return runPREditBody()
}

func runPREditBody() error {
	// Ensure gh is available
	if err := ensureGh(); err != nil {
		return err
	}

	// Get current branch
	branch, err := getCurrentBranch()
	if err != nil {
		return fmt.Errorf("get current branch: %w", err)
	}
	if branch == "" {
		return errors.New("could not determine current branch")
	}
	output.Infof(output.ModeAuto, "Current branch: %s\n", branch)

	// Fetch PR for current branch
	pr, err := getPRForCurrentBranch(branch)
	if err != nil {
		return err
	}
	if pr == nil {
		return errors.New("no open PR found for current branch. Create one first with 'gh pr create' or 'noji pr create'")
	}

	// Prepare editable buffer seeded with the current body only
	content := formatPRForEdit(pr.Body)
	tmpFile, err := createTempFile(content)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Resolve editor from config or override flag
	ed, err := config.GetEditor()
	if err != nil {
		return err
	}
	if ov := os.Getenv("NOJI_EDITOR_OVERRIDE"); strings.TrimSpace(ov) != "" {
		ed = ov
	}
	// Open in editor
	if err := openInEditor(ed, tmpFile); err != nil {
		return err
	}

	// Read back raw
	edited, err := os.ReadFile(tmpFile)
	if err != nil {
		return fmt.Errorf("read edited file: %w", err)
	}
	newBody := string(edited)

	// If body unchanged, exit early
	if newBody == pr.Body {
		output.Infof(output.ModeAuto, "No changes detected.\n")
		return nil
	}

	// Update body via gh
	if err := updatePRBody(pr.Number, newBody); err != nil {
		return err
	}

	output.Successf(output.ModeAuto, "PR #%d updated.\n", pr.Number)
	return nil
}

func runPREditTitle() error {
	// Ensure gh is available
	if err := ensureGh(); err != nil {
		return err
	}

	// Get current branch
	branch, err := getCurrentBranch()
	if err != nil {
		return fmt.Errorf("get current branch: %w", err)
	}
	if branch == "" {
		return errors.New("could not determine current branch")
	}
	output.Infof(output.ModeAuto, "Current branch: %s\n", branch)

	// Fetch PR for current branch
	pr, err := getPRForCurrentBranch(branch)
	if err != nil {
		return err
	}
	if pr == nil {
		return errors.New("no open PR found for current branch. Create one first with 'gh pr create' or 'noji pr create'")
	}

	// Seed temp file with current title + newline
	tmpFile, err := createTempFile(pr.Title + "\n")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Resolve editor from config
	ed, err := config.GetEditor()
	if err != nil {
		return err
	}
	if ov := os.Getenv("NOJI_EDITOR_OVERRIDE"); strings.TrimSpace(ov) != "" {
		ed = ov
	}
	if err := openInEditor(ed, tmpFile); err != nil {
		return err
	}

	// Read back and trim trailing newlines
	b, err := os.ReadFile(tmpFile)
	if err != nil {
		return fmt.Errorf("read edited file: %w", err)
	}
	newTitle := strings.TrimRight(string(b), "\r\n")

	if newTitle == pr.Title {
		output.Infof(output.ModeAuto, "No changes detected.\n")
		return nil
	}

	if strings.TrimSpace(newTitle) == "" {
		return errors.New("title cannot be empty")
	}

	if err := updatePRTitle(pr.Number, newTitle); err != nil {
		return err
	}
	output.Successf(output.ModeAuto, "PR #%d title updated.\n", pr.Number)
	return nil
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
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

func openInEditor(editor, filename string) error {
	// Use configured editor strictly (no env fallback here); then fallback to common ones
	candidates := []string{}
	if strings.TrimSpace(editor) != "" {
		candidates = append(candidates, editor)
	}
	candidates = append(candidates, "vim", "vi", "nvim")
	var lastErr error
	for _, ed := range candidates {
		cmd := exec.Command(ed, filename)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return fmt.Errorf("no usable editor found (tried vim, vi, nvim): last error: %v", lastErr)
}

func updatePRTitle(number int, title string) error {
	args := []string{"pr", "edit", fmt.Sprintf("%d", number), "--title", title}
	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh pr edit failed: %w", err)
	}
	return nil
}
