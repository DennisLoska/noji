package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

type Mode int

const (
	ModeAuto Mode = iota
	ModeAlways
	ModeNever
)

var (
	stdout = os.Stdout
	stderr = os.Stderr

	cInfo    = color.New(color.FgCyan, color.Bold)
	cSuccess = color.New(color.FgGreen, color.Bold)
	cWarn    = color.New(color.FgYellow, color.Bold)
	cError   = color.New(color.FgRed, color.Bold)

	currentMode = ModeAuto
)

// Init stores the chosen mode for later prints.
func Init(mode Mode) { currentMode = mode }

func parseMode(s string) (Mode, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "auto", "":
		return ModeAuto, nil
	case "always":
		return ModeAlways, nil
	case "never":
		return ModeNever, nil
	default:
		return ModeAuto, fmt.Errorf("invalid color mode: %s (want auto|always|never)", s)
	}
}

// resolveAuto decides if color should be enabled for a given writer.
func resolveAuto(w *os.File) bool {
	if force := os.Getenv("CLICOLOR_FORCE"); force == "1" {
		return true
	}
	if noColor := os.Getenv("NO_COLOR"); noColor != "" {
		return false
	}
	if clicolor := os.Getenv("CLICOLOR"); clicolor == "0" {
		return false
	}
	return term.IsTerminal(int(w.Fd()))
}

func colorEnabledFor(mode Mode, w *os.File) bool {
	switch mode {
	case ModeAlways:
		return true
	case ModeNever:
		return false
	default:
		return resolveAuto(w)
	}
}

// Writers allow tests to swap out dests.
func SetWriters(out, err *os.File) { stdout, stderr = out, err }

func Infof(_ Mode, format string, a ...any) { // mode param kept for backward compat, ignored
	printWith(stdout, cInfo, format, a...)
}

func Successf(_ Mode, format string, a ...any) {
	printWith(stdout, cSuccess, format, a...)
}

func Warnf(_ Mode, format string, a ...any) {
	printWith(stderr, cWarn, format, a...)
}

func Errorf(_ Mode, format string, a ...any) {
	printWith(stderr, cError, format, a...)
}

func Printf(_ Mode, format string, a ...any) {
	printWith(stdout, color.New(), format, a...)
}

func printWith(w *os.File, c *color.Color, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	if colorEnabledFor(currentMode, w) {
		_, _ = c.Fprintf(w, msg)
	} else {
		_, _ = fmt.Fprint(w, msg)
	}
}
