package opencode

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
func RunWithPrompt(model string, prompt string) error {
	args := []string{"run", "-m", model, prompt}
	cmd := exec.Command("opencode", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
