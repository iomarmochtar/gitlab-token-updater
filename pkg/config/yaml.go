package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v2"
)

// expandConfig expanding configuration if any of them use `include` props
func expandConfig(cfg *Config, configPath string) error {
	configDir := filepath.Dir(configPath)
	var tmpManagedTokens []ManagedToken
	for idx := range cfg.Managed {
		managed := cfg.Managed[idx]
		if managed.Ref == "" {
			managed.Ref = configPath
			tmpManagedTokens = append(tmpManagedTokens, managed)
			continue
		}
		log.Debug().Int("sequence", idx).Msg("detected include manage config")

		fpath := filepath.Join(configDir, managed.Ref)
		if !fileExists(fpath) {
			return fmt.Errorf("included file %s is not exists", fpath)
		}
		includedContent, err := os.ReadFile(filepath.Clean(fpath))
		if err != nil {
			return fmt.Errorf("error while read include file %s: %v", fpath, err)
		}
		var manageTokens []ManagedToken
		if err = yaml.Unmarshal(includedContent, &manageTokens); err != nil {
			return fmt.Errorf("error in included file %s as yaml content: %v", fpath, err)
		}

		if len(manageTokens) == 0 {
			return fmt.Errorf("included file (%s) not contains any of managed token config", fpath)
		}
		// loop it in injecting reference
		for _, mt := range manageTokens {
			mt.Ref = fpath
			tmpManagedTokens = append(tmpManagedTokens, mt)
		}
	}

	log.Debug().Interface("result", tmpManagedTokens).Int("total", len(tmpManagedTokens)).Msg("replacing manage token config")
	cfg.Managed = tmpManagedTokens
	for x, y := range tmpManagedTokens {
		log.Debug().Interface("datanya", y).Int("sequence", x).Msg("see me")
	}
	return nil
}

// ReadYAMLConfigFile read configuration in yaml content file
func ReadYAMLConfigFile(path string) (*Config, error) {
	yamlContent, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	cfg := NewConfig()
	if err = yaml.Unmarshal(yamlContent, cfg); err != nil {
		return nil, fmt.Errorf("error in unmarshal YAML object: %w", err)
	}

	if err = expandConfig(cfg, path); err != nil {
		return nil, err
	}

	if err = cfg.InitValues(); err != nil {
		return nil, err
	}

	return cfg, nil
}
