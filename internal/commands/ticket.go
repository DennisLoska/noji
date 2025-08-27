package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dennisloska/noji/internal/commands/output"
	"github.com/dennisloska/noji/internal/config"
	"github.com/dennisloska/noji/internal/opencode"
	"github.com/spf13/cobra"
)

func newTicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ticket",
		Short: "Work with tickets",
	}
	cmd.AddCommand(newTicketUpdateCmd())
	cmd.AddCommand(newTicketEditCmd())
	return cmd
}

func newTicketUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update a ticket using opencode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrompt("ticket_update.txt")
		},
	}
}

func newTicketEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <TICKET_KEY>",
		Short: "Edit a ticket description using your editor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.TrimSpace(args[0])
			if key == "" {
				return errors.New("ticket key is required")
			}
			openFlag, _ := cmd.Flags().GetBool("open")
			return runTicketEdit(key, openFlag)
		},
	}
	cmd.Flags().Bool("open", false, "open the ticket in the browser after updating")
	return cmd
}

func runTicketEdit(key string, openAfter bool) error {
	// 1) Fetch current description via opencode prompt
	model, err := config.GetModel()
	if err != nil {
		return err
	}
	promptsDir, err := config.PromptsDir()
	if err != nil {
		return err
	}
	promptPath := filepath.Join(promptsDir, "ticket_edit.txt")
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("read prompt file %s: %w", promptPath, err)
	}
	// Build prompt by appending the key as last line instruction
	prompt := string(promptBytes) + "\nTicket key: " + key + "\n"

	// Capture opencode output to a buffer rather than streaming to stdout
	desc, err := runOpencodeCapture(model, prompt)
	if err != nil {
		return err
	}

	// 2) Open editor with the description
	tmpFile, err := createTempFile(desc)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

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

	edited, err := os.ReadFile(tmpFile)
	if err != nil {
		return fmt.Errorf("read edited file: %w", err)
	}
	newDesc := string(edited)
	if newDesc == desc {
		output.Infof(output.ModeAuto, "No changes detected.\n")
		return nil
	}

	// 3) Write back via opencode using MCP to update the ticket Description exactly
	delimStart := "---BEGIN_DESCRIPTION---"
	delimEnd := "---END_DESCRIPTION---"
	updatePrompt := fmt.Sprintf("Use only the Atlassian MCP server tools (no web). Replace the Jira issue %s Description field with EXACTLY the content between %s and %s. Do not add, remove, rephrase, or format anything.\n%s\n%s\n%s", key, delimStart, delimEnd, delimStart, newDesc, delimEnd)
	if err := opencode.RunWithPrompt(model, updatePrompt); err != nil {
		return err
	}

	output.Successf(output.ModeAuto, "Ticket %s description updated.\n", key)
	if openAfter {
		if err := openTicketInBrowser(key); err != nil {
			output.Infof(output.ModeAuto, "Could not open browser: %v\n", err)
		}
	}
	return nil
}

// runOpencodeCapture runs opencode and returns its stdout as string.
func openTicketInBrowser(key string) error {
	// Try to construct a Jira URL from environment or git remote
	// First, allow explicit base via NOJI_JIRA_BASE like https://jira.example.com/browse
	if base := strings.TrimRight(os.Getenv("NOJI_JIRA_BASE"), "/"); base != "" {
		return openURL(base + "/" + key)
	}
	// Fallback: ask Atlassian MCP for the browse URL for this key
	model, err := config.GetModel()
	if err != nil {
		return err
	}
	prompt := fmt.Sprintf("Using only Atlassian MCP tools (no web), return ONLY the direct browser URL to open the Jira issue %s (no extra text).", key)
	url, err := runOpencodeCapture(model, prompt)
	if err != nil {
		return err
	}
	url = strings.TrimSpace(url)
	if url == "" {
		return errors.New("could not resolve ticket URL")
	}
	return openURL(url)
}

func openURL(u string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", u).Run()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Run()
	default:
		if _, err := exec.LookPath("xdg-open"); err == nil {
			return exec.Command("xdg-open", u).Run()
		}
		return fmt.Errorf("no method to open URL on %s", runtime.GOOS)
	}
}

func runOpencodeCapture(model, prompt string) (string, error) {
	args := []string{"run", "-m", model, prompt}
	// We cannot reuse opencode.RunWithPrompt because it streams to stdout
	cmd := exec.Command("opencode", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("opencode run failed: %w", err)
	}
	return out.String(), nil
}
