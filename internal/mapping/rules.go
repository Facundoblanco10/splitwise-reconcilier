package mapping

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Member struct {
	SplitwiseID   int    `yaml:"splitwise_id"`
	DefaultEntity string `yaml:"default_entity"`
}

type Config struct {
	Splitwise struct {
		GroupID  int `yaml:"group_id"`
		MyUserID int `yaml:"my_user_id"`
	} `yaml:"splitwise"`
	Members []Member `yaml:"members"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	return &cfg, nil
}
