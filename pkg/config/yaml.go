package config

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// ReadYAMLConfig read yaml config formatted content
func ReadYAMLConfig(yamlContent []byte) (*Config, error) {
	var err error
	cfg := NewConfig()
	if err = yaml.Unmarshal(yamlContent, cfg); err != nil {
		return nil, fmt.Errorf("error in unmarshal YAML object: %w", err)
	}

	if err = cfg.InitValues(); err != nil {
		return nil, err
	}

	return cfg, nil
}
