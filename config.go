package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Config paths relative to user home
const (
	configDirName = ".traycli"
	commandFile   = "command.txt"
	stdoutFile    = "stdout.txt"
	stderrFile    = "stderr.txt"
)

// Config holds paths for traycli configuration and output files.
type Config struct {
	CommandPath string
	StdoutPath  string
	StderrPath  string
}

// LoadConfig returns config paths. Returns error if home directory cannot be determined.
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, configDirName)
	return &Config{
		CommandPath: filepath.Join(dir, commandFile),
		StdoutPath:  filepath.Join(dir, stdoutFile),
		StderrPath:  filepath.Join(dir, stderrFile),
	}, nil
}

// ReadCommand reads the command from command.txt. Returns empty string if file does not exist
// or content is empty/whitespace.
func ReadCommand(cfg *Config) (string, error) {
	data, err := os.ReadFile(cfg.CommandPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	cmd := strings.TrimSpace(string(data))
	return cmd, nil
}
