package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config paths relative to user home
const (
	configDirName = ".traycli"
	configFile    = "config.json"
	stdoutFile    = "stdout.txt"
	stderrFile    = "stderr.txt"
)

// Config holds paths for traycli configuration and output files.
type Config struct {
	ConfigPath string
	StdoutPath string
	StderrPath string
}

// CommandConfig is the parsed content of config.json.
type CommandConfig struct {
	Cmd []string          `json:"cmd"`
	Env map[string]string `json:"env,omitempty"`
}

// LoadConfig returns config paths. Returns error if home directory cannot be determined.
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, configDirName)
	return &Config{
		ConfigPath: filepath.Join(dir, configFile),
		StdoutPath: filepath.Join(dir, stdoutFile),
		StderrPath: filepath.Join(dir, stderrFile),
	}, nil
}

// ReadConfig reads and parses config.json. Returns (nil, nil) if file does not exist.
func ReadConfig(cfg *Config) (*CommandConfig, error) {
	data, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cc CommandConfig
	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, err
	}
	return &cc, nil
}

// WriteEmptyConfig creates the config directory and writes an empty config template.
func WriteEmptyConfig(cfg *Config) error {
	dir := filepath.Dir(cfg.ConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data := []byte(`{"cmd": [], "env": {}}`)
	return os.WriteFile(cfg.ConfigPath, data, 0644)
}
