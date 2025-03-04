package app_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/iomarmochtar/gitlab-token-updater/app"
	cfg "github.com/iomarmochtar/gitlab-token-updater/pkg/config"
	gl "github.com/iomarmochtar/gitlab-token-updater/pkg/gitlab"
	t_helper "github.com/iomarmochtar/gitlab-token-updater/test"
	gm "github.com/iomarmochtar/gitlab-token-updater/test/mocks/gitlab"
	sm "github.com/iomarmochtar/gitlab-token-updater/test/mocks/shell"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGitlabTokenUpdater_Do(t *testing.T) {
	testCases := map[string]struct {
		config         func() *cfg.Config
		mockGitlab     func(*gomock.Controller) *gm.MockGitlabAPI
		mockShell      func(*gomock.Controller) *sm.MockShell
		forceNew       bool
		dryRun         bool
		strict         bool
		currentTime    *time.Time
		envVars        map[string]string
		expectedErrMsg string
	}{
		"success: simple scenario": {
			config: func() *cfg.Config {
				c := t_helper.GenConfig(nil, nil, nil)
				return c
			},
			currentTime: t_helper.GenTime("2024-04-05"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				newToken := "glpat-newnew"
				g := gm.NewMockGitlabAPI(ctrl)

				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				g.EXPECT().RotateRepoToken(t_helper.SampleRepoPath, 123, *t_helper.GenTime("2024-07-04")).Return(newToken, nil)
				g.EXPECT().UpdateRepoVar(t_helper.SampleRepoPath, t_helper.SampleCICDVar, newToken).Return(nil)

				return g
			},
			mockShell: func(ctrl *gomock.Controller) *sm.MockShell {
				return nil
			},
		},
		"update access token that will expired in 1 month ahead and execute all hooks": {
			envVars: map[string]string{
				"SUBS1": "value1",
			},
			config: func() *cfg.Config {
				anotherManageTokens := t_helper.GenManageTokens(nil, nil, nil)
				anotherManageTokens[0].Type = cfg.ManagedTypeGroup
				anotherManageTokens[0].Path = t_helper.SampleGroupPath
				anotherManageTokens[0].Tokens[0].Hooks[0] = t_helper.SampleHookUpdateVarGroup

				c := t_helper.GenConfig(anotherManageTokens, nil, nil)
				c.Managed[0].Tokens[0].Hooks = append(c.Managed[0].Tokens[0].Hooks, cfg.Hook{
					Type: cfg.HookTypeExecCMD,
					Args: map[string]any{
						"path": t_helper.SamplePathToScript,
						"env": map[any]any{
							"ENV1": "additional_env",
							"ENV2": "subs-${SUBS1}",
						},
					},
				})
				return c
			},
			currentTime: t_helper.GenTime("2024-04-05"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				newToken := "glpat-newnew"
				g := gm.NewMockGitlabAPI(ctrl)

				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				g.EXPECT().RotateRepoToken(t_helper.SampleRepoPath, 123, *t_helper.GenTime("2024-07-04")).Return(newToken, nil)
				g.EXPECT().UpdateRepoVar(t_helper.SampleRepoPath, t_helper.SampleCICDVar, newToken).Return(nil)

				groupAccessTokens := []gl.GitlabAccessToken{t_helper.SampleGroupAccessToken}
				g.EXPECT().ListGroupAccessToken(t_helper.SampleGroupPath).Return(groupAccessTokens, nil)
				g.EXPECT().RotateGroupToken(t_helper.SampleGroupPath, 123, *t_helper.GenTime("2024-07-04")).Return(newToken, nil)
				g.EXPECT().UpdateGroupVar(t_helper.SampleGroupPath, t_helper.SampleCICDVar, newToken).Return(nil)
				return g
			},
			mockShell: func(ctrl *gomock.Controller) *sm.MockShell {
				s := sm.NewMockShell(ctrl)
				expectedEnvVar := map[string]string{"GL_NEW_TOKEN": "glpat-newnew", "ENV1": "additional_env", "ENV2": "subs-value1"}
				s.EXPECT().Exec(t_helper.SamplePathToScript, expectedEnvVar).Return([]byte("abc"), nil)
				return s
			},
		},
		"skip: no access token will be expired and no hooks executed": {
			config: func() *cfg.Config {
				return t_helper.GenConfig(nil, nil, nil)
			},
			currentTime: t_helper.GenTime("2024-01-01"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}
				g := gm.NewMockGitlabAPI(ctrl)
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
		},
		"skip: access token without expiry excluded from the execution": {
			config: func() *cfg.Config {
				return t_helper.GenConfig(nil, nil, nil)
			},
			currentTime: t_helper.GenTime("2024-01-01"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				accessTokens := []gl.GitlabAccessToken{
					{
						Name:      t_helper.SampleAccessTokeName,
						Type:      gl.GitlabTargetTypeRepo,
						ID:        123,
						Path:      t_helper.SampleRepoPath,
						Active:    true,
						Revoked:   false,
						ExpiresAt: nil,
					},
				}
				g := gm.NewMockGitlabAPI(ctrl)
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
		},
		"any errors occured in the middle execution will not break the iterrations but will be raised in the end": {
			config: func() *cfg.Config {
				// there are 4 managed
				var additionalsMT []cfg.ManagedToken
				additionalsMT = append(additionalsMT, t_helper.GenManageTokens(
					t_helper.GenManageTokens(
						t_helper.GenManageTokens(nil, nil, nil),
						nil, nil),
					nil, nil)...)
				c := t_helper.GenConfig(additionalsMT, nil, nil)
				c.Managed[0].Path = "/first"
				c.Managed[1].Path = "/second"
				c.Managed[2].Path = "/third"
				c.Managed[3].Path = "/fourth"

				// set the first managed's hook as script execution
				c.Managed[0].Tokens[0].Hooks[0] = t_helper.SampleHookExecScript

				// no hooks set in 4th
				c.Managed[3].Tokens[0].Hooks = nil
				return c
			},
			currentTime: t_helper.GenTime("2024-04-05"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				newToken := "glpat-newnew"
				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}

				g := gm.NewMockGitlabAPI(ctrl)

				// the first iter running normally
				g.EXPECT().ListRepoAccessToken("/first").Return(accessTokens, nil).Times(1)
				g.EXPECT().RotateRepoToken(t_helper.SampleRepoPath, 123, *t_helper.GenTime("2024-07-04")).Return(newToken, nil).Times(1)

				// the second iter there was an error in during renew token, and it's causing no hook been executed
				g.EXPECT().ListRepoAccessToken("/second").Return(accessTokens, nil).Return(accessTokens, nil).Times(1)
				g.EXPECT().RotateRepoToken(t_helper.SampleRepoPath, 123, *t_helper.GenTime("2024-07-04")).Return("", fmt.Errorf("error during renew")).Times(1)

				// the third iter, there error happen in listing access token
				g.EXPECT().ListRepoAccessToken("/third").Return(accessTokens, nil).Return(nil, fmt.Errorf("error in listing access token")).Times(1)

				// the fourth iter running normally and no hook executed
				g.EXPECT().ListRepoAccessToken("/fourth").Return(accessTokens, nil).Times(1)
				g.EXPECT().RotateRepoToken(t_helper.SampleRepoPath, 123, *t_helper.GenTime("2024-07-04")).Return(newToken, nil).Times(1)

				return g
			},
			mockShell: func(ctrl *gomock.Controller) *sm.MockShell {
				// hook exec for the first iter access token renew
				s := sm.NewMockShell(ctrl)
				s.EXPECT().Exec(t_helper.SamplePathToScript, map[string]string{"GL_NEW_TOKEN": "glpat-newnew"}).Return([]byte("abc"), nil)
				return s
			},
			expectedErrMsg: "some error(s) occured during execution",
		},
		"force new mode enabled then all of access tokens are forced to be renewed": {
			config: func() *cfg.Config {
				return t_helper.GenConfig(nil, nil, nil)
			},
			currentTime: t_helper.GenTime("2024-01-01"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				newToken := "glpat-newnew"
				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}
				g := gm.NewMockGitlabAPI(ctrl)
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				g.EXPECT().RotateRepoToken(t_helper.SampleRepoPath, 123, *t_helper.GenTime("2024-03-31")).Return(newToken, nil)
				g.EXPECT().UpdateRepoVar(t_helper.SampleRepoPath, t_helper.SampleCICDVar, newToken).Return(nil)
				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
			forceNew: true,
		},
		"revoked access token will be skipped": {
			config: func() *cfg.Config {
				return t_helper.GenConfig(nil, nil, nil)
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				accessTokens := []gl.GitlabAccessToken{
					{
						Name:      t_helper.SampleAccessTokeName,
						ID:        123,
						Path:      t_helper.SampleRepoPath,
						Revoked:   true,
						ExpiresAt: t_helper.GenTime("2024-05-01"),
					},
				}
				g := gm.NewMockGitlabAPI(ctrl)
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
		},
		"inactive access token will be skipped": {
			config: func() *cfg.Config {
				return t_helper.GenConfig(nil, nil, nil)
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				accessTokens := []gl.GitlabAccessToken{
					{
						Name:      t_helper.SampleAccessTokeName,
						ID:        123,
						Path:      t_helper.SampleRepoPath,
						Active:    false,
						ExpiresAt: t_helper.GenTime("2024-05-01"),
					},
				}
				g := gm.NewMockGitlabAPI(ctrl)
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
		},
		"dry_run: no any modify execution will be made": {
			config: func() *cfg.Config {
				anotherManageTokens := t_helper.GenManageTokens(nil, nil, nil)
				anotherManageTokens[0].Type = cfg.ManagedTypeGroup
				anotherManageTokens[0].Tokens[0].Hooks[0] = t_helper.SampleHookUpdateVarGroup

				return t_helper.GenConfig(anotherManageTokens, nil, nil)
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}
				g := gm.NewMockGitlabAPI(ctrl)
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				g.EXPECT().GetRepoVar(t_helper.SampleRepoPath, t_helper.SampleCICDVar).Return(&gl.GitlabCICDVar{}, nil)
				g.EXPECT().ListGroupAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				g.EXPECT().GetGroupVar(t_helper.SampleGroupPath, t_helper.SampleCICDVar).Return(&gl.GitlabCICDVar{}, nil)

				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
			dryRun: true,
		},
		"strict: if found an error then it will not continue to next step/iterration": {
			config: func() *cfg.Config {
				return t_helper.GenConfig(nil, nil, nil)
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				g := gm.NewMockGitlabAPI(ctrl)
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(nil, fmt.Errorf("an error while list access token"))
				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
			strict:         true,
			expectedErrMsg: "an error while list access token",
		},
		"strict: no access token found": {
			// in the strict mode, if the access token not found then it wouldn't continue
			config: func() *cfg.Config {
				return t_helper.GenConfig(nil, nil, nil)
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				g := gm.NewMockGitlabAPI(ctrl)
				accessTokens := []gl.GitlabAccessToken{{
					Name:      "",
					Type:      gl.GitlabTargetTypeRepo,
					ID:        100,
					Active:    true,
					Revoked:   false,
					ExpiresAt: t_helper.GenTime("2024-05-01"),
				}}
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
			strict:         true,
			expectedErrMsg: "token MR Handler in /path/to/repo is not exists",
		},
		"strict: in dry run mode hook exec cmd will check for script existance": {
			config: func() *cfg.Config {
				c := t_helper.GenConfig(nil, nil, nil)
				c.Managed[0].Tokens[0].Hooks[0] = t_helper.SampleHookExecScript
				return c
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}
				g := gm.NewMockGitlabAPI(ctrl)
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				return g
			},
			mockShell: func(ctrl *gomock.Controller) *sm.MockShell {
				s := sm.NewMockShell(ctrl)
				s.EXPECT().FileMustExists(t_helper.SamplePathToScript).Return(fmt.Errorf("script not exists"))
				return s
			},
			strict:         true,
			dryRun:         true,
			expectedErrMsg: "script not exists",
		},
		"strict: error in renew process": {
			currentTime: t_helper.GenTime("2024-04-05"),
			config: func() *cfg.Config {
				// there are 2 managed
				c := t_helper.GenConfig(t_helper.GenManageTokens(nil, nil, nil), nil, nil)
				c.Managed[0].Path = "/first"
				c.Managed[1].Path = "/second"

				return c
			},
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}

				g := gm.NewMockGitlabAPI(ctrl)
				// the first iter running normally
				g.EXPECT().ListRepoAccessToken("/first").Return(accessTokens, nil).Return(accessTokens, nil)
				g.EXPECT().RotateRepoToken(t_helper.SampleRepoPath, 123, *t_helper.GenTime("2024-07-04")).Return("", fmt.Errorf("error during renew"))

				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
			strict:         true,
			expectedErrMsg: "error during renew",
		},
		"hook: if retry and can be recovered then no error will be returned": {
			config: func() *cfg.Config {
				c := t_helper.GenConfig(nil, nil, nil)
				c.Managed[0].Tokens[0].Hooks[0] = cfg.Hook{
					Type:  cfg.HookTypeUpdateVar,
					Retry: 1,
					Args: map[string]any{
						"name": t_helper.SampleCICDVar,
						"path": t_helper.SampleRepoPath,
						"type": cfg.ManagedTypeRepository,
					},
				}
				return c
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				newToken := "glpat-newnew"
				g := gm.NewMockGitlabAPI(ctrl)

				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				g.EXPECT().RotateRepoToken(t_helper.SampleRepoPath, 123, *t_helper.GenTime("2024-07-27")).Return(newToken, nil)
				g.EXPECT().UpdateRepoVar(t_helper.SampleRepoPath, t_helper.SampleCICDVar, newToken).Return(fmt.Errorf("error occured")).Times(1)
				g.EXPECT().UpdateRepoVar(t_helper.SampleRepoPath, t_helper.SampleCICDVar, newToken).Return(nil).Times(1)

				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
		},
		"hook: the CICD var is located in external Gitlab instance": {
			config: func() *cfg.Config {
				hookInternal := []cfg.Hook{
					t_helper.SampleHookUpdateVarGroup,
					{
						Type: cfg.HookTypeUpdateVar,
						Args: map[string]any{
							"type":         cfg.ManagedTypeGroup,
							"name":         t_helper.SampleCICDVar,
							"path":         t_helper.SampleGroupPath,
							"gitlab":       "https://another2.gitlab.dev",
							"gitlab_token": "glpat-another2",
						},
					},
				}
				c := t_helper.GenConfig(nil, nil, hookInternal)

				// set in the first hook as indicating the Gitlab api caller is switched to another instance
				c.Managed[0].Tokens[0].Hooks[0] = cfg.Hook{
					Type: cfg.HookTypeUpdateVar,
					Args: map[string]any{
						"type":         cfg.ManagedTypeRepository,
						"name":         t_helper.SampleCICDVar,
						"path":         t_helper.SampleRepoPath,
						"gitlab":       t_helper.SampleAnotherGitlab,
						"gitlab_token": t_helper.SampleAnotherGitlabToken,
					},
				}
				return c
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				newToken := "glpat-newnew"
				g := gm.NewMockGitlabAPI(ctrl)
				anotherGL := gm.NewMockGitlabAPI(ctrl)
				anotherGL2 := gm.NewMockGitlabAPI(ctrl)

				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				g.EXPECT().RotateRepoToken(t_helper.SampleRepoPath, 123, *t_helper.GenTime("2024-07-27")).Return(newToken, nil)
				g.EXPECT().UpdateGroupVar(t_helper.SampleGroupPath, t_helper.SampleCICDVar, newToken).Return(nil).Times(1)
				g.EXPECT().InitGitlab(t_helper.SampleAnotherGitlab, t_helper.SampleAnotherGitlabToken).Return(anotherGL, nil).Times(1)
				g.EXPECT().InitGitlab("https://another2.gitlab.dev", "glpat-another2").Return(anotherGL2, nil).Times(1)

				// updating each target
				anotherGL.EXPECT().UpdateRepoVar(t_helper.SampleRepoPath, t_helper.SampleCICDVar, newToken).Return(nil).Times(1)
				anotherGL2.EXPECT().UpdateGroupVar(t_helper.SampleGroupPath, t_helper.SampleCICDVar, newToken).Return(nil).Times(1)

				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
		},
		"hook: the CICD var is located in external Gitlab instance in dry run mode": {
			dryRun: true,
			config: func() *cfg.Config {
				hookInternal := []cfg.Hook{
					t_helper.SampleHookUpdateVarGroup,
					{
						Type: cfg.HookTypeUpdateVar,
						Args: map[string]any{
							"type":         cfg.ManagedTypeGroup,
							"name":         t_helper.SampleCICDVar,
							"path":         t_helper.SampleGroupPath,
							"gitlab":       "https://another2.gitlab.dev",
							"gitlab_token": "glpat-another2",
						},
					},
				}
				c := t_helper.GenConfig(nil, nil, hookInternal)

				// set in the first hook as indicating the Gitlab api caller is switched to another instance
				c.Managed[0].Tokens[0].Hooks[0] = cfg.Hook{
					Type: cfg.HookTypeUpdateVar,
					Args: map[string]any{
						"type":         cfg.ManagedTypeRepository,
						"name":         t_helper.SampleCICDVar,
						"path":         t_helper.SampleRepoPath,
						"gitlab":       t_helper.SampleAnotherGitlab,
						"gitlab_token": t_helper.SampleAnotherGitlabToken,
					},
				}
				return c
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				g := gm.NewMockGitlabAPI(ctrl)
				anotherGL := gm.NewMockGitlabAPI(ctrl)
				anotherGL2 := gm.NewMockGitlabAPI(ctrl)

				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				g.EXPECT().GetGroupVar(t_helper.SampleGroupPath, t_helper.SampleCICDVar).Return(&gl.GitlabCICDVar{}, nil).Times(1)
				g.EXPECT().InitGitlab(t_helper.SampleAnotherGitlab, t_helper.SampleAnotherGitlabToken).Return(anotherGL, nil).Times(1)
				g.EXPECT().InitGitlab("https://another2.gitlab.dev", "glpat-another2").Return(anotherGL2, nil).Times(1)

				// updating each target
				anotherGL.EXPECT().GetRepoVar(t_helper.SampleRepoPath, t_helper.SampleCICDVar).Return(&gl.GitlabCICDVar{}, nil).Times(1)
				anotherGL2.EXPECT().GetGroupVar(t_helper.SampleGroupPath, t_helper.SampleCICDVar).Return(&gl.GitlabCICDVar{}, nil).Times(1)

				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
		},
		"hook: failed in update_var with external Gitlab set": {
			config: func() *cfg.Config {
				c := t_helper.GenConfig(nil, nil, nil)
				c.Managed[0].Tokens[0].Hooks[0] = cfg.Hook{
					Type: cfg.HookTypeUpdateVar,
					Args: map[string]any{
						"type":         cfg.ManagedTypeRepository,
						"name":         t_helper.SampleCICDVar,
						"path":         t_helper.SampleRepoPath,
						"gitlab":       t_helper.SampleAnotherGitlab,
						"gitlab_token": t_helper.SampleAnotherGitlabToken,
					},
				}
				return c
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				newToken := "glpat-newnew"
				g := gm.NewMockGitlabAPI(ctrl)

				accessTokens := []gl.GitlabAccessToken{t_helper.SampleRepoAccessToken}
				g.EXPECT().ListRepoAccessToken(t_helper.SampleRepoPath).Return(accessTokens, nil)
				g.EXPECT().RotateRepoToken(t_helper.SampleRepoPath, 123, *t_helper.GenTime("2024-07-27")).Return(newToken, nil)
				g.EXPECT().InitGitlab(t_helper.SampleAnotherGitlab, t_helper.SampleAnotherGitlabToken).Return(nil, fmt.Errorf("got an error during initiating Gitlab")).Times(1)

				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
			expectedErrMsg: "some error(s) occured during execution",
		},
		"manage_personal: using hook use_token": {
			config: func() *cfg.Config {
				c := cfg.NewConfig()
				c.Token = "glpat-abc"
				c.Managed = []cfg.ManagedToken{
					{
						Type: cfg.ManagedTypePersonal,
						Tokens: []cfg.AccessToken{
							{
								Name: "personal_pat_1",
								Hooks: []cfg.Hook{
									{
										Type: cfg.HookTypeUseToken,
									},
									t_helper.SampleHookUpdateVarRepo,
								},
							},
						},
					},
				}

				return c
			},
			currentTime: t_helper.GenTime("2024-04-28"),
			mockGitlab: func(ctrl *gomock.Controller) *gm.MockGitlabAPI {
				newToken := "glpat-newnew"
				g := gm.NewMockGitlabAPI(ctrl)

				accessTokens := []gl.GitlabAccessToken{t_helper.SamplePersonalAccessToken}
				g.EXPECT().ListPersonalAccessToken().Return(accessTokens, nil)
				g.EXPECT().RotatePersonalToken(123, *t_helper.GenTime("2024-07-27")).Return(newToken, nil)
				g.EXPECT().Auth(newToken).Return(nil)
				g.EXPECT().UpdateRepoVar(t_helper.SampleRepoPath, t_helper.SampleCICDVar, newToken).Return(nil).Times(1)

				return g
			},
			mockShell: func(*gomock.Controller) *sm.MockShell {
				return nil
			},
		},
	}

	for title, tc := range testCases {
		t.Run(title, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			if tc.envVars != nil {
				for k, v := range tc.envVars {
					defer os.Unsetenv(k)
					_ = os.Setenv(k, v)
				}
			}

			config := tc.config()
			assert.NoError(t, config.InitValues())

			updater := app.NewGitlabTokenUpdater(
				config,
				tc.mockGitlab(ctrl),
				tc.mockShell(ctrl),
			).
				WithDryRun(tc.dryRun).
				WithForceRenew(tc.forceNew).
				WithStrictMode(tc.strict)

			if tc.currentTime != nil {
				updater.WithCustomCurrentTime(tc.currentTime)
			}

			err := updater.Do()

			if tc.expectedErrMsg != "" {
				assert.EqualError(t, err, tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
