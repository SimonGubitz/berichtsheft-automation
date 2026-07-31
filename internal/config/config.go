package config

import (
	"errors"
)

var ErrNoConfigFile = errors.New("no such config file")

type Config struct {
	Username     string `json:"username"`
	JiraHost     string `json:"jiraHost"`
	SystemPrompt string `json:"systemPrompt,omitempty"`
}

type ConfigClient interface {
	GetConfig() (*Config, error)
	SetConfig(conf Config) error
}
