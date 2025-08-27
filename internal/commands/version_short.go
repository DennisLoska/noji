package commands

import (
	"fmt"

	"github.com/dennis/noji/internal/version"
)

func mustShortVersion() string {
	v := version.Version
	if v == "" || v == "dev" {
		return "dev"
	}
	if v[0] != 'v' {
		return fmt.Sprintf("v%s", v)
	}
	return v
}
