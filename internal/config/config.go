package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/adrg/xdg"
)

var ErrNoConfigFile = errors.New("no such config file")

type Config struct {
	Username     string `json:"username"`
	JiraUrl      string `json:"jiraUrl"`
	SystemPrompt string `json:"systemPrompt,omitempty"`
}

func GetConfig() (*Config, error) {
	configFilePath, err := xdg.ConfigFile("berichtsheft/config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to get config file path. %w", err)
	}

	f, err := os.OpenFile(configFilePath, os.O_RDONLY, 0600)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoConfigFile
		}
		return nil, fmt.Errorf("failed to open config file at %s. %w", configFilePath, err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal contents of file %s. %w", configFilePath, err)
	}

	return &config, nil
}

func SetConfig(config Config) error {
	configFilePath, err := xdg.ConfigFile("berichtsheft/config.json")
	if err != nil {
		return fmt.Errorf("failed to get config file path. %w", err)
	}

	f, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open config file at %s. %w", configFilePath, err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	err = encoder.Encode(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config into file %s. %w", configFilePath, err)
	}

	return nil
}
