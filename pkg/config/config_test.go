package config_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	c "github.com/iomarmochtar/gitlab-token-updater/pkg/config"
	"github.com/stretchr/testify/assert"
)

var (
	sampleHookExecScript = c.Hook{
		Type: c.HookTypeExecCMD,
		Args: map[string]string{
			"path": "./path/to/script.sh",
		},
	}
	sampleHookUpdateVar = c.Hook{
		Type: c.HookTypeUpdateVar,
		Args: map[string]string{
			"name": "SOME_VAR",
			"path": "/path/to/repo",
			"type": c.ManagedTypeRepository,
		},
	}
)

func genSampleManagedTokens() []c.ManagedToken {
	return []c.ManagedToken{
		{
			Path: "/path/go/repo",
			Type: c.ManagedTypeRepository,
			Tokens: []c.AccessToken{
				{
					Name:        "TF_IaC",
					RenewBefore: "3d",
					Hooks:       []c.Hook{sampleHookExecScript},
				},
			},
		},
	}
}

func TestConfig_InitValues_Validations(t *testing.T) {
	testCases := map[string]struct {
		Cfg         func() *c.Config
		ExpectedErr error
	}{
		"ok": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "glpat-abc"
				cfg.Managed = genSampleManagedTokens()
				return cfg
			},
			ExpectedErr: nil,
		},
		"empty token list": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Managed = []c.ManagedToken{}
				return cfg
			},
			ExpectedErr: c.ErrValidationEmptyManagedList,
		},
		"invalid default renew before": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "glpat-abc"
				cfg.DefaultRenewBefore = "1m"
				cfg.Managed = genSampleManagedTokens()
				return cfg
			},
			ExpectedErr: c.ErrValidationInvalidDefaultRenewBefore,
		},
		"invalid default expiry after rotate": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "glpat-abc"
				cfg.DefaultExpiryAfterRotate = "1m"
				cfg.Managed = genSampleManagedTokens()
				return cfg
			},
			ExpectedErr: c.ErrValidationInvalidDefaultExpiryAfterRotate,
		},
		"empty Gitlab token": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Managed = genSampleManagedTokens()
				return cfg
			},
			ExpectedErr: c.ErrValidationEmptyGitlabToken,
		},
		"empty host": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Host = ""
				return cfg
			},
			ExpectedErr: c.ErrValidationEmptyHost,
		},
		"empty path in managed token": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "",
						Type: c.ManagedTypeRepository,
						Tokens: []c.AccessToken{
							{
								Name:        "TF_IaC",
								RenewBefore: "3d",
								Hooks:       []c.Hook{sampleHookExecScript},
							},
						},
					},
				}

				return cfg
			},
			ExpectedErr: c.ErrValidationManagedEmptyPath,
		},
		"invalid type": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: "wrong_type",
						Tokens: []c.AccessToken{
							{
								Name:        "TF_IaC",
								RenewBefore: "3d",
								Hooks:       []c.Hook{sampleHookExecScript},
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationManagedInvalidType,
		},
		"empty managed token list": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: c.ManagedTypeRepository,
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationManagedEmptyTokenList,
		},
		"empty managed token name": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: c.ManagedTypeRepository,
						Tokens: []c.AccessToken{
							{
								Name: "",
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationTokenEmptyName,
		},
		"invalid renew before value": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: c.ManagedTypeRepository,
						Tokens: []c.AccessToken{
							{
								Name:        "abc",
								RenewBefore: "10p",
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationManagedInvalidRenewBefore,
		},
		"invalid expiry after rotate value": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: c.ManagedTypeRepository,
						Tokens: []c.AccessToken{
							{
								Name:              "abc",
								ExpiryAfterRotate: "10m",
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationManagedInvalidExpiryAfterRotate,
		},
		"invalid hook type": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: c.ManagedTypeRepository,
						Tokens: []c.AccessToken{
							{
								Name: "TF IaC",
								Hooks: []c.Hook{
									{
										Type: "wrong_type",
									},
								},
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationHookInvalidType,
		},
		"invalid type hook update_var": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: c.ManagedTypeRepository,
						Tokens: []c.AccessToken{
							{
								Name: "TF IaC",
								Hooks: []c.Hook{
									{
										Type: c.HookTypeUpdateVar,
										Args: map[string]string{
											"name": "SOME_VAR",
											"path": "/path/to/repo",
											"type": "wrong_type",
										},
									},
								},
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationHookUpdateVarInvalidType,
		},

		"hook update_var missing name": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: c.ManagedTypeRepository,
						Tokens: []c.AccessToken{
							{
								Name: "TF IaC",
								Hooks: []c.Hook{
									{
										Type: c.HookTypeUpdateVar,
										Args: map[string]string{
											"path": "/some/path",
										},
									},
								},
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationHookUpdateVarMissingName,
		},
		"hook update_var missing path": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: c.ManagedTypeRepository,
						Tokens: []c.AccessToken{
							{
								Name: "TF IaC",
								Hooks: []c.Hook{
									{
										Type: c.HookTypeUpdateVar,
										Args: map[string]string{
											"name": "CI_VAR",
										},
									},
								},
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationHookUpdateVarMissingPath,
		},
		"hook exec_cmd missing path": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: c.ManagedTypeRepository,
						Tokens: []c.AccessToken{
							{
								Name: "TF IaC",
								Hooks: []c.Hook{
									{
										Type: c.HookTypeExecCMD,
									},
								},
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationHookExecCMDMissingPath,
		},
		"hook use_token: only for manage_token personal": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Path: "/some/path",
						Type: c.ManagedTypeGroup,
						Tokens: []c.AccessToken{
							{
								Name: "access_token_ro",
								Hooks: []c.Hook{
									{
										Type: c.HookTypeUseToken,
									},
								},
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationHookUseTokenNotByPersonalType,
		},
		"hook use_token: can be only use once": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Type: c.ManagedTypePersonal,
						Tokens: []c.AccessToken{
							{
								Name: "access_token_ro",
								Hooks: []c.Hook{
									{
										Type: c.HookTypeUseToken,
									},
								},
							},
							{
								Name: "access_token_rw",
								Hooks: []c.Hook{
									{
										Type: c.HookTypeUseToken,
									},
								},
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationHookUseTokenAlreadyUse,
		},
		"hook use_token: must be set at the first sequence": {
			Cfg: func() *c.Config {
				cfg := c.NewConfig()
				cfg.Token = "abc"
				cfg.Managed = []c.ManagedToken{
					{
						Type: c.ManagedTypePersonal,
						Tokens: []c.AccessToken{
							{
								Name: "TF IaC",
								Hooks: []c.Hook{
									{
										Type: c.HookTypeExecCMD,
										Args: map[string]string{
											"path": "./some/path",
										},
									},
									{
										Type: c.HookTypeUseToken,
									},
								},
							},
						},
					},
				}
				return cfg
			},
			ExpectedErr: c.ErrValidationHookUseTokenNotFirstSeq,
		},
	}

	for title, tc := range testCases {
		t.Run(title, func(t *testing.T) {
			cfg := tc.Cfg()
			res := cfg.InitValues()
			assert.ErrorIs(t, res, tc.ExpectedErr, fmt.Sprintf("expecting returning error %v ~> %s", tc.ExpectedErr, res))
		})
	}
}

func TestConfig_InitValues_Defaults(t *testing.T) {
	t.Run("token config env var subs", func(t *testing.T) {
		envName := "GITLAB_TOKEN"
		defer os.Unsetenv(envName)
		_ = os.Setenv(envName, "abc")

		cfg := c.NewConfig()
		cfg.Managed = genSampleManagedTokens()
		cfg.Token = "glpat-${GITLAB_TOKEN}"

		assert.NoError(t, cfg.InitValues())
		assert.Equal(t, "glpat-abc", cfg.Token)
	})

	t.Run("host config env var subs", func(t *testing.T) {
		envName := "GITLAB_HOST"
		defer os.Unsetenv(envName)
		_ = os.Setenv(envName, "git.repo.internal")

		cfg := c.NewConfig()
		cfg.Host = "https://${GITLAB_HOST}"
		cfg.Token = "glpat-abc"
		cfg.Managed = genSampleManagedTokens()

		assert.NoError(t, cfg.InitValues())
		assert.Equal(t, "https://git.repo.internal", cfg.Host)
	})

	t.Run("config env var is not set results to blank", func(t *testing.T) {
		cfg := c.NewConfig()
		cfg.Token = "token-${NO_ENV_VAR}"
		cfg.Managed = genSampleManagedTokens()

		assert.NoError(t, cfg.InitValues())
		assert.Equal(t, "token-", cfg.Token)
	})
}

func TestDuration_Funcs(t *testing.T) {
	propList := map[any][]string{
		&c.Config{}: {
			"DefaultRenewBefore",
			"DefaultExpiryAfterRotate",
		},
		&c.AccessToken{}: {
			"RenewBefore",
			"ExpiryAfterRotate",
		},
	}

	for obj, props := range propList {
		for idx := range props {
			prop := props[idx]
			methodName := fmt.Sprintf("%sDuration", prop)
			refl := reflect.ValueOf(obj).Elem()
			field := refl.FieldByName(prop)

			t.Run("invalid pattern", func(t *testing.T) {
				field.Set(reflect.ValueOf("abc"))
				method := refl.MethodByName(methodName)
				results := method.Call(nil)

				dur := results[0].Interface().(time.Duration)
				err := results[1].Interface().(error)

				assert.Equal(t, time.Duration(0), dur)
				assert.ErrorContains(t, err, "abc is not match with duration pattern", "invalid")
			})

			t.Run("ok", func(t *testing.T) {
				field.Set(reflect.ValueOf("3M"))
				method := refl.MethodByName(methodName)
				results := method.Call(nil)

				dur := results[0].Interface().(time.Duration)
				err := results[1].Interface()

				expected := (3 * 30 * 24) * time.Hour

				assert.Nil(t, err)
				assert.Equal(t, expected.Hours(), dur.Hours())
			})
		}
	}
}

func TestHook_StrArgs(t *testing.T) {
	t.Run("hook update var", func(t *testing.T) {
		assert.Equal(t,
			"type:repository,path:/path/to/repo,name:SOME_VAR",
			sampleHookUpdateVar.StrArgs())
	})

	t.Run("hook exec cmd", func(t *testing.T) {
		assert.Equal(t,
			"path:./path/to/script.sh",
			sampleHookExecScript.StrArgs())
	})

	t.Run("use_token hook", func(t *testing.T) {
		assert.Equal(t, "", c.Hook{}.StrArgs())
	})
}

func TestNewConfig_Defaults(t *testing.T) {
	obj := c.NewConfig()

	assert.Equal(t, "https://gitlab.com/", obj.Host)
	assert.Equal(t, "${GL_RENEWER_TOKEN}", obj.Token)
	assert.Equal(t, "14d", obj.DefaultRenewBefore)
	assert.Equal(t, "3M", obj.DefaultExpiryAfterRotate)
	assert.Equal(t, uint8(0), obj.DefaultHookRetry)
}
