// Package app main package in executing gitlab access token logic
package app

import (
	"errors"
	"time"

	cfg "github.com/iomarmochtar/gitlab-token-updater/pkg/config"
	gl "github.com/iomarmochtar/gitlab-token-updater/pkg/gitlab"
	"github.com/iomarmochtar/gitlab-token-updater/pkg/shell"
	"github.com/rs/zerolog/log"
)

const (
	injectEnvVarShellExec = "GL_NEW_TOKEN"
	dryRunDommyToken      = "glpat-abc"
)

var (
	ErrDuringExecution = errors.New("some error(s) occured during execution")
)

type accessTokenPair struct {
	glAccessToken  gl.GitlabAccessToken
	cfgAccessToken cfg.AccessToken
}

// GitlabTokenUpdater hold required properties and main execution of gitlab-token-updater
type GitlabTokenUpdater struct {
	config     *cfg.Config
	sh         shell.Shell
	glAPI      gl.GitlabAPI
	now        *time.Time
	forceRenew bool
	dryRun     bool
	strict     bool
	errors     []error
}

func (g GitlabTokenUpdater) listAccessTokens(mg cfg.ManagedToken) (results []accessTokenPair, err error) {
	var tokens []gl.GitlabAccessToken
	switch mg.Type {
	case cfg.ManagedTypeRepository:
		tokens, err = g.glAPI.ListRepoAccessToken(mg.Path)
	case cfg.ManagedTypeGroup:
		tokens, err = g.glAPI.ListGroupAccessToken(mg.Path)
	case cfg.ManagedTypePersonal:
		tokens, err = g.glAPI.ListPersonalAccessToken()
	}
	if err != nil {
		return nil, err
	}

	for _, token := range mg.Tokens {
		isFound := false
		for _, scToken := range tokens {
			// ignore the revoked one
			if scToken.Revoked {
				tkID := scToken.ID
				log.Debug().
					Str("token", scToken.Path).
					Str("path", scToken.Path).
					Msgf("access token by id %d is revoked, skip it", tkID)
				continue
			}

			if scToken.Name == token.Name {
				results = append(results, accessTokenPair{
					glAccessToken:  scToken,
					cfgAccessToken: token,
				})
				isFound = true
				break
			}
		}

		if !isFound {
			log.Warn().Str("token", token.Name).Str("path", mg.Path).Msg("token is not found")
		}
	}

	return results, nil
}

func (g GitlabTokenUpdater) processRenew(tkn accessTokenPair) (string, error) {
	if g.dryRun {
		return dryRunDommyToken, nil
	}

	path := tkn.glAccessToken.Path
	id := tkn.glAccessToken.ID
	nextExpiration, _ := tkn.cfgAccessToken.ExpiryAfterRotateDuration()
	nextExpiry := g.now.Add(nextExpiration)
	if tkn.glAccessToken.Type == gl.GitlabTargetTypePersonal {
		return g.glAPI.RotatePersonalToken(id, nextExpiry)
	} else if tkn.glAccessToken.Type == gl.GitlabTargetTypeRepo {
		return g.glAPI.RotateRepoToken(path, id, nextExpiry)
	}

	return g.glAPI.RotateGroupToken(path, id, nextExpiry)
}

func (g GitlabTokenUpdater) execHook(hk cfg.Hook, newToken string) (err error) {
	switch hk.Type {
	case cfg.HookTypeUseToken:
		if g.dryRun {
			return nil
		}
		return g.glAPI.Auth(newToken)
	case cfg.HookTypeUpdateVar:
		args := hk.UpdateVarArgs()
		if args.Type == cfg.ManagedTypeRepository {
			if g.dryRun {
				_, err := g.glAPI.GetRepoVar(args.Path, args.Name)
				return err
			}

			return g.glAPI.UpdateRepoVar(
				args.Path,
				args.Name,
				newToken,
			)
		}

		if g.dryRun {
			_, err := g.glAPI.GetGroupVar(args.Path, args.Name)
			return err
		}

		return g.glAPI.UpdateGroupVar(
			args.Path,
			args.Name,
			newToken,
		)
	case cfg.HookTypeExecCMD:
		args := hk.ExecCMDArgs()
		if g.dryRun {
			return g.sh.FileMustExists(args.Path)
		}

		results, err := g.sh.Exec(args.Path, map[string]string{injectEnvVarShellExec: newToken})
		log.Debug().Msgf("script execution results %s", string(results))
		return err
	}

	return nil
}

// errAppender appending error if not in strict mode, otherwise just return the error as is
func (g *GitlabTokenUpdater) errAppender(err error) error {
	if err != nil && !g.strict {
		g.errors = append(g.errors, err)
		return nil
	}
	return err
}

// Do the main sequences of app logic
func (g *GitlabTokenUpdater) Do() error {
	for _, mg := range g.config.Managed {
		path := mg.Path
		logPath := log.With().Str("path", path).Str("m_type", mg.Type).Logger()
		logPath.Info().Msg("processing")

		ats, err := g.listAccessTokens(mg)
		if err != nil {
			logPath.Error().Err(err).Msg("error while listing access token")
			if err = g.errAppender(err); err != nil {
				return err
			}
			continue
		}

		for _, at := range ats {
			logTkn := logPath.With().Str("token", at.cfgAccessToken.Name).Logger()
			logTkn.Info().Msg("processing")

			befDur, _ := at.cfgAccessToken.RenewBeforeDuration()
			addMe := g.now.Add(befDur)
			expiredAt := *at.glAccessToken.ExpiresAt
			validToRenew := false
			if addMe.After(expiredAt) {
				logTkn.Warn().Msgf("reach renew time. expired: %v, renew before: %s", expiredAt, at.cfgAccessToken.RenewBefore)
				validToRenew = true
			}

			if !(g.forceRenew || validToRenew) {
				logTkn.Debug().Msg("not identified as need to renew")
				continue
			}

			logTkn.Info().Msg("processing token renewal")
			newToken, err := g.processRenew(at)
			if err != nil {
				logTkn.Error().Err(err).Msg("error renew token")
				if err = g.errAppender(err); err != nil {
					return err
				}
				continue
			}
			logTkn.Info().Msg("token successfully renewed")

			if len(at.cfgAccessToken.Hooks) < 1 {
				logTkn.Debug().Msg("no hook configured")
				continue
			}
			logTkn.Info().Msg("executing hooks")

			for _, hk := range at.cfgAccessToken.Hooks {
				logHook := logTkn.With().Str("hook_type", hk.Type).Str("args", hk.StrArgs()).Logger()
				var lastErr error
				for i := 1; i <= int(hk.Retry+1); i++ {
					logHookAttempt := logHook.With().Int("attempt", i).Logger()

					logHookAttempt.Debug().Msg("executing hook")
					err = g.execHook(hk, newToken)
					if err == nil {
						logHookAttempt.Info().Msg("hook successfully executed")
						lastErr = nil
						break
					}

					logHookAttempt.Error().Err(err).Msg("error in hook execution")
					lastErr = err
				}
				if err = g.errAppender(lastErr); err != nil {
					return err
				}
			}
		}
	}

	log.Info().Msg("done")
	if len(g.errors) == 0 {
		return nil
	}

	log.Warn().Msg("detected following error(s) during execution")
	for idErr := range g.errors {
		errExec := g.errors[idErr]
		log.Warn().Int("sequence", idErr+1).Msgf("- %v", errExec)
	}
	return ErrDuringExecution
}

// WithForceRenew set enble/disable force renew scenario
func (g *GitlabTokenUpdater) WithForceRenew(e bool) *GitlabTokenUpdater {
	g.forceRenew = e
	return g
}

// WithDryRun set enable/disable dry run mode
func (g *GitlabTokenUpdater) WithDryRun(e bool) *GitlabTokenUpdater {
	g.dryRun = e
	return g
}

// WithStrictMode set enable/disable strict mode
func (g *GitlabTokenUpdater) WithStrictMode(e bool) *GitlabTokenUpdater {
	g.strict = e
	return g
}

// WithCustomCurrentTime set custom current time, used in test
func (g *GitlabTokenUpdater) WithCustomCurrentTime(tm *time.Time) *GitlabTokenUpdater {
	g.now = tm
	return g
}

// NewGitlabTokenUpdater create GitlabTokenUpdater with it's default values
func NewGitlabTokenUpdater(config *cfg.Config, glAPI gl.GitlabAPI, sh shell.Shell) *GitlabTokenUpdater {
	now := time.Now()
	return &GitlabTokenUpdater{
		config:     config,
		glAPI:      glAPI,
		sh:         sh,
		now:        &now,
		dryRun:     false,
		forceRenew: false,
		errors:     []error{},
	}
}
