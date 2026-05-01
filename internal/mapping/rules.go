package mapping

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Card struct {
	ID          string  `yaml:"id"`
	Name        string  `yaml:"name"`
	ClosingDay  *int    `yaml:"closing_day"`
}

type Config struct {
	Splitwise struct {
		GroupID    int `yaml:"group_id"`
		MyUserID   int `yaml:"my_user_id"`
	} `yaml:"splitwise"`
	Cards               []Card         `yaml:"cards"`
	RulesByCategory     map[string]string `yaml:"rules_by_category"`
	OverridesByExpenseID map[int]string   `yaml:"overrides_by_expense_id"`
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
