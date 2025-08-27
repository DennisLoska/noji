package output

import (
	"github.com/charmbracelet/glamour"
)

// RenderMarkdown renders GitHub-flavored Markdown to ANSI, with a compact theme.
// If rendering fails, it returns the original input.
func RenderMarkdown(s string) string {
	if s == "" {
		return s
	}
	// Use a compact style; we aim for minimal padding.
	// glamour supports built-in styles: dark, light, notty, etc. There isn't an official
	// tokyonight style bundled. We'll pick dark and keep it readable.
	// If a tokyonight theme becomes available, swap style here.
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEnvironmentConfig(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return s
	}
	out, err := r.Render(s)
	if err != nil {
		return s
	}
	return out
}
