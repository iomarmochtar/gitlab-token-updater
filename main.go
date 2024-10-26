package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/urfave/cli/v2"

	"github.com/iomarmochtar/gitlab-token-updater/app"
	cfg "github.com/iomarmochtar/gitlab-token-updater/pkg/config"
	gl "github.com/iomarmochtar/gitlab-token-updater/pkg/gitlab"
	"github.com/iomarmochtar/gitlab-token-updater/pkg/shell"
)

var (
	// CmdName name of command line
	CmdName = "gitlab-token-updater"
	// Version app version, this will be injected/modified during compilation time
	Version = "0.0.0"
	// BuildHash git commit hash during build process
	BuildHash = "0000000000000000000000000000000000000000"
)

// New return command line instance in parsing and executing main instance
func New() *cli.App {
	cli.VersionPrinter = func(ctx *cli.Context) {
		_, _ = fmt.Fprintf(ctx.App.Writer, `{"version": "%s", "commit": "%s", "compile_time": "%v"}`,
			ctx.App.Version, BuildHash, ctx.App.Compiled)
	}
	cmd := &cli.App{
		Name: CmdName,
		Authors: []*cli.Author{
			{
				Name:  "Imam Omar Mochtar",
				Email: "iomarmochtar@gmail.com",
			},
		},
		Usage:   "Gitlab repository and group access token updater/renewal",
		Version: Version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Enable debug mode",
				EnvVars: []string{"DEBUG_MODE"},
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "all token will be updated regardless the specified config",
			},
			&cli.BoolFlag{
				Name:    "strict",
				Aliases: []string{"s"},
				Usage:   "enable strict mode, if any of error found the it will raise the errors",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "dry run mode, skip any write execution",
			},
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "path of yaml config path",
				Required: true,
			},
		},
		Before: func(ctx *cli.Context) error {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			if ctx.Bool("debug") {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
				zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
			}
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

			return nil
		},
		Action: func(ctx *cli.Context) error {
			configPath := ctx.String("config")
			forceRenew := ctx.Bool("force")
			dryRun := ctx.Bool("dry-run")
			strictMode := ctx.Bool("strict")

			yamlContent, err := os.ReadFile(filepath.Clean(configPath))
			if err != nil {
				return err
			}

			config, err := cfg.ReadYAMLConfig(yamlContent)
			if err != nil {
				return err
			}

			glAPI, err := gl.NewGitlabAPI(config.Host, config.Token)
			if err != nil {
				return err
			}

			if forceRenew {
				log.Warn().Msg("force renew enabled")
			}

			if dryRun {
				log.Warn().Msg("dry run mode enabled")
			}

			if strictMode {
				log.Warn().Msg("strict mode enabled")
			}

			return app.
				NewGitlabTokenUpdater(config, glAPI, &shell.SHExecutor{}).
				WithDryRun(dryRun).
				WithForceRenew(forceRenew).
				WithStrictMode(strictMode).
				Do()
		},
	}
	return cmd
}

func main() {
	a := New()
	if err := a.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
