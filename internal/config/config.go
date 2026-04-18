package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Anthropic   AnthropicConfig   `json:"anthropic"`
	Box         BoxConfig         `json:"box"`
	Local       LocalConfig       `json:"local"`
	Defaults    DefaultsConfig    `json:"defaults"`
	Preferences PreferencesConfig `json:"preferences"`
}

type AnthropicConfig struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
}

type BoxConfig struct {
	MCPServerURL string `json:"mcp_server_url"`
	MCPAuthToken string `json:"mcp_auth_token"`
	RootFolderID string `json:"root_folder_id"`
}

type LocalConfig struct {
	SyncFolder   string `json:"sync_folder"`
	ConfigFolder string `json:"config_folder"`
}

type DefaultsConfig struct {
	SettingGenre string `json:"setting_genre"`
	Mode         string `json:"mode"` // "series" or "standalone"
}

type PreferencesConfig struct {
	FastDraftDefault       bool `json:"fast_draft_default"`
	AutoAdvancePhase       bool `json:"auto_advance_phase"`
	SceneWordTargetDefault int  `json:"scene_word_target_default"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = "config.json"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg.Local.SyncFolder = expandHome(cfg.Local.SyncFolder)
	cfg.Local.ConfigFolder = expandHome(cfg.Local.ConfigFolder)
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Anthropic.Model == "" {
		c.Anthropic.Model = "claude-sonnet-4-6"
	}
	if c.Anthropic.MaxTokens == 0 {
		c.Anthropic.MaxTokens = 8192
	}
	if c.Local.SyncFolder == "" {
		c.Local.SyncFolder = "~/writing-harness-sync"
	}
	if c.Local.ConfigFolder == "" {
		c.Local.ConfigFolder = "~/.writing-harness"
	}
	if c.Defaults.SettingGenre == "" {
		c.Defaults.SettingGenre = "Sci-Fi"
	}
	if c.Defaults.Mode == "" {
		c.Defaults.Mode = "series"
	}
	if c.Preferences.SceneWordTargetDefault == 0 {
		c.Preferences.SceneWordTargetDefault = 2000
	}
}

func (c *Config) validate() error {
	if c.Box.MCPServerURL == "" {
		return fmt.Errorf("box.mcp_server_url is required in config.json")
	}
	return nil
}

func expandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
