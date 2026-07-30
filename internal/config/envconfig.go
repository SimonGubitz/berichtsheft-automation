package config

import (
	"fmt"
	"os"
)

type EnvConfig struct{}

func (ec *EnvConfig) GetConfig() (*Config, error) {

	username := os.Getenv("JIRA_USERNAME")
	jiraHost := os.Getenv("JIRA_URL")

	return &Config{
		Username: username,
		JiraHost: jiraHost,
	}, nil
}

func (ec *EnvConfig) SetConfig(conf Config) error {

	err := os.Setenv("JIRA_USERNAME", conf.Username)
	if err != nil {
		return fmt.Errorf("failed to set JIRA_USERNAME. %w", err)
	}

	err = os.Setenv("JIRA_URL", conf.JiraHost)
	if err != nil {
		return fmt.Errorf("failed to set JIRA_URL. %w", err)
	}

	return nil
}
