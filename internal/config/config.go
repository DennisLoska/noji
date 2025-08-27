package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	appDirName = "noji"
	configName = "config"
	configType = "yaml"
	keyModel   = "model"
	keyEditor  = "editor"
	promptsDir = "prompts"
)

// EnsureConfig sets up config dir, default config, and placeholder prompts.
// Resolution order:
// 1) If NOJI_CONFIG_HOME is set, use $NOJI_CONFIG_HOME/noji
// 2) Else use os.UserConfigDir()/noji (platform-correct; honors XDG_CONFIG_HOME)
func EnsureConfig() (configPath string, promptsPath string, err error) {
	configHome := os.Getenv("NOJI_CONFIG_HOME")
	if configHome == "" {
		var derr error
		configHome, derr = os.UserConfigDir()
		if derr != nil {
			return "", "", fmt.Errorf("get user config dir: %w", derr)
		}
	}
	appDir := filepath.Join(configHome, appDirName)
	prompts := filepath.Join(appDir, promptsDir)
	if err := os.MkdirAll(prompts, 0o755); err != nil {
		return "", "", fmt.Errorf("create config dirs: %w", err)
	}

	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath(appDir)
	v.SetDefault(keyModel, "github-copilot/gpt-4.1")
	v.SetDefault(keyEditor, "vim")

	cfgFile := filepath.Join(appDir, configName+"."+configType)
	if _, statErr := os.Stat(cfgFile); errors.Is(statErr, os.ErrNotExist) {
		if err := v.WriteConfigAs(cfgFile); err != nil {
			return "", "", fmt.Errorf("write default config: %w", err)
		}
	}

	// Seed prompt files from repository templates if missing (never overwrite)
	repoBase := "prompts"
	templates := []string{"pr_create.txt", "pr_update.txt", "ticket_update.txt", "ticket_edit.txt"}
	for _, name := range templates {
		userPath := filepath.Join(prompts, name)
		if _, err := os.Stat(userPath); errors.Is(err, os.ErrNotExist) {
			repoPath := filepath.Join(repoBase, name)
			if b, rerr := os.ReadFile(repoPath); rerr == nil {
				_ = os.WriteFile(userPath, b, 0o644)
			} else {
				// Fallback to empty file if repo template missing
				_ = os.WriteFile(userPath, []byte(""), 0o644)
			}
		}
	}

	return cfgFile, prompts, nil
}

// resolveAppDir returns the app directory respecting NOJI_CONFIG_HOME override.
func resolveAppDir() (string, error) {
	configHome := os.Getenv("NOJI_CONFIG_HOME")
	if configHome == "" {
		var err error
		configHome, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(configHome, appDirName), nil
}

// GetModel reads the selected model from config.
func GetModel() (string, error) {
	if _, _, err := EnsureConfig(); err != nil {
		return "", err
	}
	v := viper.New()
	appDir, err := resolveAppDir()
	if err != nil {
		return "", err
	}
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath(appDir)
	if err := v.ReadInConfig(); err != nil {
		return "", err
	}
	return v.GetString(keyModel), nil
}

// GetEditor reads the preferred editor from config (defaults to vim).
func GetEditor() (string, error) {
	if _, _, err := EnsureConfig(); err != nil {
		return "", err
	}
	v := viper.New()
	appDir, err := resolveAppDir()
	if err != nil {
		return "", err
	}
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath(appDir)
	if err := v.ReadInConfig(); err != nil {
		return "", err
	}
	ed := v.GetString(keyEditor)
	if strings.TrimSpace(ed) == "" {
		return "vim", nil
	}
	return ed, nil
}

// SetModel writes the selected model to config.
func SetModel(model string) error {
	if _, _, err := EnsureConfig(); err != nil {
		return err
	}
	v := viper.New()
	appDir, err := resolveAppDir()
	if err != nil {
		return err
	}
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath(appDir)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	v.Set(keyModel, model)
	return v.WriteConfig()
}

// SetEditor writes the preferred editor to config.
func SetEditor(editor string) error {
	if _, _, err := EnsureConfig(); err != nil {
		return err
	}
	v := viper.New()
	appDir, err := resolveAppDir()
	if err != nil {
		return err
	}
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath(appDir)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	v.Set(keyEditor, editor)
	return v.WriteConfig()
}

// PromptsDir returns the prompts directory path.
func PromptsDir() (string, error) {
	_, p, err := EnsureConfig()
	return p, err
}
