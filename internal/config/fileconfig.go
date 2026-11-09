package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

type FileConfig struct{}

func NewFileConfig() *FileConfig {
	return &FileConfig{}
}

func (fc *FileConfig) GetConfig() (*Config, error) {
	configFilePath, err := xdg.ConfigFile(filepath.Join("berichtsheft", "config.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to get config file path. %w", err)
	}

	f, err := os.OpenFile(configFilePath, os.O_RDONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoConfigFile
		}
		return nil, fmt.Errorf("failed to open config file at %s. %w", configFilePath, err)
	}
	defer func() { _ = f.Close() }()

	decoder := json.NewDecoder(f)
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal contents of file %s. %w", configFilePath, err)
	}

	return &config, nil
}

func (fc *FileConfig) SetConfig(config Config) error {
	configFilePath, err := xdg.ConfigFile(filepath.Join("berichtsheft", "config.json"))
	if err != nil {
		return fmt.Errorf("failed to get config file path. %w", err)
	}

	f, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open config file at %s. %w", configFilePath, err)
	}
	defer func() { _ = f.Close() }()

	encoder := json.NewEncoder(f)
	err = encoder.Encode(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config into file %s. %w", configFilePath, err)
	}

	return nil
}
