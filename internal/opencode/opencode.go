package opencode

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dennis/noji/internal/commands/output"
)

// ListModels runs `opencode models` and returns the list.
func ListModels() ([]string, error) {
	cmd := exec.Command("opencode", "models")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("opencode models failed: %w; output: %s", err, out.String())
	}
	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	var models []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			models = append(models, line)
		}
	}
	if len(models) == 0 {
		return nil, errors.New("no models returned by opencode")
	}
	return models, nil
}

// RunWithPrompt runs: opencode run -m <model> "<prompt>"
func RunWithPrompt(model string, prompt string, markdown bool) error {
	args := []string{"run", "-m", model, prompt}
	cmd := exec.Command("opencode", args...)

	if markdown {
		// Capture output for markdown rendering
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		cmd.Stdin = os.Stdin

		err := cmd.Run()
		if err != nil {
			// Still output stderr even on error
			if stderrBuf.Len() > 0 {
				os.Stderr.Write(stderrBuf.Bytes())
			}
			return err
		}

		// Render stdout as markdown and output
		if stdoutBuf.Len() > 0 {
			rendered := output.RenderMarkdown(stdoutBuf.String())
			os.Stdout.WriteString(rendered)
		}

		// Output stderr as-is
		if stderrBuf.Len() > 0 {
			os.Stderr.Write(stderrBuf.Bytes())
		}

		return nil
	} else {
		// Normal output mode
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
}
