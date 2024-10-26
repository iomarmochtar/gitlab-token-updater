package config_test

import (
	"os"
	"testing"

	c "github.com/iomarmochtar/gitlab-token-updater/pkg/config"
	"github.com/stretchr/testify/assert"
)

var (
	yamlConfigBasic = `
token: glpat-abc
default_renew_before: 2M
default_hook_retry: 1
manage_tokens:
  - path: path/to/repo
    type: repository
    access_tokens:
      - name: TF IaC
        renew_before: 3M
        hooks:
          - type: exec_cmd
            args:
              path: ./path/to/cmd
`

	yamlConfigEnvVar = `
token: ${TEST_GL_TOKEN}
default_expiry_after_rotate: 2M
manage_tokens:
  - path: path/to/repo
    type: repository
    access_tokens:
      - name: TF IaC`
)

func TestReadYAMLConfig(t *testing.T) {
	t.Run("read basic configuration", func(t *testing.T) {
		cfg, err := c.ReadYAMLConfig([]byte(yamlConfigBasic))

		assert.NotNil(t, cfg)
		assert.NoError(t, err)
		assert.Equal(t, "3M", cfg.Managed[0].Tokens[0].RenewBefore, "overrided value")
		assert.Equal(t, uint8(1), cfg.Managed[0].Tokens[0].Hooks[0].Retry, "use the default value")
	})

	t.Run("subs var in value", func(t *testing.T) {
		varName := "TEST_GL_TOKEN"
		defer os.Unsetenv(varName)
		_ = os.Setenv(varName, "glpat-abc")
		cfg, err := c.ReadYAMLConfig([]byte(yamlConfigEnvVar))

		assert.NotNil(t, cfg)
		assert.NoError(t, err)
		assert.Equal(t, "glpat-abc", cfg.Token)
		assert.Equal(t, "14d", cfg.Managed[0].Tokens[0].RenewBefore, "default renew before if not defined")
		assert.Equal(t, "2M", cfg.Managed[0].Tokens[0].ExpiryAfterRotate, "default value of expirty after rotate if not defined")
	})

	t.Run("invalid yaml content", func(t *testing.T) {
		cfg, err := c.ReadYAMLConfig([]byte(`not a valid one`))

		assert.Nil(t, cfg)
		assert.Error(t, err)
	})

	t.Run("missing required config", func(t *testing.T) {
		cfg, err := c.ReadYAMLConfig([]byte(`token: glpat-abc`))

		assert.Nil(t, cfg)
		assert.ErrorIs(t, err, c.ErrValidationEmptyManagedList)
	})
}
