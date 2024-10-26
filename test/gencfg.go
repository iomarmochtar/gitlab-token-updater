// Package test collections of test helpers
package test

import (
	"time"

	cfg "github.com/iomarmochtar/gitlab-token-updater/pkg/config"
	gl "github.com/iomarmochtar/gitlab-token-updater/pkg/gitlab"
)

var (
	SampleAccessToken    = "glpat-abc"
	SampleRepoPath       = "/path/to/repo"
	SampleGroupPath      = "/path/to/group"
	SampleCICDVar        = "SOME_VAR"
	SamplePathToScript   = "./path/to/script.sh"
	SampleHookExecScript = cfg.Hook{
		Type: cfg.HookTypeExecCMD,
		Args: map[string]string{
			"path": SamplePathToScript,
		},
	}
	SampleHookUpdateVarRepo = cfg.Hook{
		Type: cfg.HookTypeUpdateVar,
		Args: map[string]string{
			"name": SampleCICDVar,
			"path": SampleRepoPath,
			"type": cfg.ManagedTypeRepository,
		},
	}
	SampleHookUpdateVarGroup = cfg.Hook{
		Type: cfg.HookTypeUpdateVar,
		Args: map[string]string{
			"name": SampleCICDVar,
			"path": SampleGroupPath,
			"type": cfg.ManagedTypeGroup,
		},
	}
	SamplePersonalAccessToken = gl.GitlabAccessToken{
		Name:      "personal_pat_1",
		Type:      gl.GitlabTargetTypePersonal,
		ID:        123,
		Active:    true,
		Revoked:   false,
		ExpiresAt: GenTime("2024-05-01"),
	}
	SampleRepoAccessToken = gl.GitlabAccessToken{
		Name:      "MR Handler",
		Type:      gl.GitlabTargetTypeRepo,
		ID:        123,
		Path:      SampleRepoPath,
		Active:    true,
		Revoked:   false,
		ExpiresAt: GenTime("2024-05-01"),
	}
	SampleGroupAccessToken = gl.GitlabAccessToken{
		Name:      "MR Handler",
		Type:      gl.GitlabTargetTypeGroup,
		ID:        123,
		Path:      SampleGroupPath,
		Active:    true,
		Revoked:   false,
		ExpiresAt: GenTime("2024-05-01"),
	}
)

// GenTime generate time.Time instance based on given date string pattern
func GenTime(dt string) *time.Time {
	layout := "2006-01-02"

	date, _ := time.Parse(layout, dt)
	return &date
}

// GenHooks generate hook config instance
func GenHooks(addHks ...cfg.Hook) []cfg.Hook {
	hks := []cfg.Hook{SampleHookUpdateVarRepo}
	hks = append(hks, addHks...)

	return hks
}

// GenAccessTokens generate access token config instance
func GenAccessTokens(addTks []cfg.AccessToken, addHks []cfg.Hook) []cfg.AccessToken {
	ats := []cfg.AccessToken{
		{
			Name:        "MR Handler",
			RenewBefore: "1M",
			Hooks:       GenHooks(addHks...),
		},
	}
	ats = append(ats, addTks...)

	return ats
}

// GenManageTokens generate manage tokens config instance
func GenManageTokens(addMTs []cfg.ManagedToken, addTks []cfg.AccessToken, addHks []cfg.Hook) []cfg.ManagedToken {
	manageTokens := []cfg.ManagedToken{
		{
			Type:   cfg.ManagedTypeRepository,
			Path:   SampleRepoPath,
			Tokens: GenAccessTokens(addTks, addHks),
		},
	}
	manageTokens = append(manageTokens, addMTs...)

	return manageTokens
}

// GenConfig generate main config instance
func GenConfig(addMTs []cfg.ManagedToken, addTks []cfg.AccessToken, addHks []cfg.Hook) *cfg.Config {
	c := cfg.NewConfig()
	c.Managed = GenManageTokens(addMTs, addTks, addHks)
	c.Token = SampleAccessToken
	return c
}
