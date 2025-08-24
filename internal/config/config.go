package config

import (
	"os"

	"github.com/leogtzr/payment-dates-advisor/internal/model"
	"gopkg.in/yaml.v3"
)

// Load loads and parses the YAML configuration file
func Load(path string) (model.Config, error) {
	var cfg model.Config
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
