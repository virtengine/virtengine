package coordinator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	configFile     = "config.json"
	transcriptFile = "transcript.json"
)

type State struct {
	BaseDir string
}

func (s State) ConfigPath() string {
	return filepath.Join(s.BaseDir, configFile)
}

func (s State) TranscriptPath() string {
	return filepath.Join(s.BaseDir, transcriptFile)
}

func (s State) Phase1Dir() string {
	return filepath.Join(s.BaseDir, "phase1")
}

func (s State) Phase2Dir() string {
	return filepath.Join(s.BaseDir, "phase2")
}

func (s State) EnsureDirs() error {
	for _, dir := range []string{s.Phase1Dir(), s.Phase2Dir()} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create dir %s: %w", dir, err)
		}
	}
	return nil
}

func (s State) LoadConfig() (*Config, error) {
	data, err := os.ReadFile(s.ConfigPath())
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s State) SaveConfig(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.ConfigPath(), data, 0o600)
}

func (s State) LoadTranscript() ([]byte, error) {
	return os.ReadFile(s.TranscriptPath())
}

func (s State) SaveTranscript(data []byte) error {
	return os.WriteFile(s.TranscriptPath(), data, 0o600)
}

func (s State) Phase1ContributionPaths() ([]string, error) {
	return listContributionFiles(s.Phase1Dir())
}

func (s State) Phase2ContributionPaths() ([]string, error) {
	return listContributionFiles(s.Phase2Dir())
}

func listContributionFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) == 0 {
			continue
		}
		if filepath.Ext(name) != ".bin" {
			continue
		}
		if len(name) >= len("contrib-0000.bin") && name[:7] == "contrib" {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)
	return files, nil
}
