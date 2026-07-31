package config

type VarConfig struct {
	username string
	jiraHost string
}

func NewVarConfig(username, jiraHost string) *VarConfig {
	return &VarConfig{
		username: username,
		jiraHost: jiraHost,
	}
}

func (vc *VarConfig) GetConfig() (*Config, error) {
	return &Config{
		Username: vc.username,
		JiraHost: vc.jiraHost,
	}, nil
}

func (vc *VarConfig) SetConfig(conf Config) error {
	vc.username = conf.Username
	vc.jiraHost = conf.JiraHost
	return nil
}
