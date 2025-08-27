package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	appDirName = "noji"
	configName = "config"
	configType = "yaml"
	keyModel   = "model"
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
	v.SetDefault(keyModel, "github-copilot/gpt-5")

	cfgFile := filepath.Join(appDir, configName+"."+configType)
	if _, statErr := os.Stat(cfgFile); errors.Is(statErr, os.ErrNotExist) {
		if err := v.WriteConfigAs(cfgFile); err != nil {
			return "", "", fmt.Errorf("write default config: %w", err)
		}
	}

	// Ensure placeholder prompt files exist
	placeholders := map[string]string{
		"pr_create.txt":     "Placeholder prompt for PR create",
		"pr_update.txt":     "Placeholder prompt for PR update",
		"ticket_update.txt": "Placeholder prompt for ticket update",
	}
	for name, content := range placeholders {
		p := filepath.Join(prompts, name)
		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			_ = os.WriteFile(p, []byte(content+"\n"), 0o644)
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

// PromptsDir returns the prompts directory path.
func PromptsDir() (string, error) {
	_, p, err := EnsureConfig()
	return p, err
}
