package config_test

import (
	"fmt"
	"os"
	"testing"

	c "github.com/iomarmochtar/gitlab-token-updater/pkg/config"
	t_helper "github.com/iomarmochtar/gitlab-token-updater/test"
	"github.com/stretchr/testify/assert"
)

type KV map[string]string

func TestReadYAMLConfigFile(t *testing.T) {
	testCases := map[string]struct {
		fixture      string
		expectedErr  string
		extraAsserts func(*testing.T, *c.Config)
		envSet       KV
	}{
		"ok: read basic config": {
			fixture: "basic_config.yml",
			extraAsserts: func(t *testing.T, cfg *c.Config) {
				assert.Equal(t, "3M", cfg.Managed[0].Tokens[0].RenewBefore, "overrided value")
				assert.Equal(t, uint8(1), cfg.Managed[0].Tokens[0].Hooks[0].Retry, "use the default value")
			},
		},
		"ok: subs var in value": {
			fixture: "config_with_subs_env.yml",
			envSet:  KV{"TEST_GL_TOKEN": "glpat-abc"},
			extraAsserts: func(t *testing.T, cfg *c.Config) {
				assert.Equal(t, "glpat-abc", cfg.Token)
				assert.Equal(t, "14d", cfg.Managed[0].Tokens[0].RenewBefore, "default renew before if not defined")
				assert.Equal(t, "2M", cfg.Managed[0].Tokens[0].ExpiryAfterRotate, "default value of expirty after rotate if not defined")
			},
		},
		"ok: test example config": {
			fixture: "../../../examples/main_config.yml",
			envSet:  KV{"GL_RENEWER_TOKEN": "glpat-abc"},
			extraAsserts: func(t *testing.T, cfg *c.Config) {
				assert.Equal(t, "glpat-abc", cfg.Token)
				assert.Equal(t, 4, len(cfg.Managed))
				assert.Equal(t, uint8(2), cfg.Managed[0].Tokens[0].Hooks[1].Retry)
			},
		},
		"err: validation check": {
			fixture:     "config_with_subs_env.yml",
			expectedErr: "empty gitlab token",
		},
		"err: invalid yaml content": {
			fixture:     "broken_config.yml",
			expectedErr: "error in unmarshal YAML object: yaml: line 2: mapping values are not allowed in this context",
		},
		"err: included file not exists": {
			fixture:     "broken_include_not_exists.yml",
			expectedErr: fmt.Sprintf("included file %s is not exists", t_helper.FixturePath("configs", "not_exists.yml")),
		},
		"err: broken included file": {
			fixture:     "broken_include_wrong_yaml_content.yml",
			expectedErr: fmt.Sprintf("error in included file %s as yaml content: yaml: line 2: mapping values are not allowed in this context", t_helper.FixturePath("configs", "broken_config.yml")),
		},
		"err: included file not contains valid manage token": {
			fixture:     "broken_include_empty_managed_token.yml",
			expectedErr: fmt.Sprintf("included file (%s) not contains any of managed token config", t_helper.FixturePath("configs", "empty.yml")),
		},
	}

	for title, tc := range testCases {
		t.Run(title, func(t *testing.T) {
			for k, v := range tc.envSet {
				_ = os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			cfg, err := c.ReadYAMLConfigFile(t_helper.FixturePath("configs", tc.fixture))

			if tc.expectedErr == "" {
				assert.NotNil(t, cfg)
				assert.NoError(t, err)
				if tc.extraAsserts != nil {
					tc.extraAsserts(t, cfg)
				}
			} else {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
			}
		})
	}
}
