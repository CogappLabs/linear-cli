package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CogappLabs/linear-cli/internal/commands"
	"github.com/CogappLabs/linear-cli/internal/config"
	"github.com/urfave/cli/v3"
)

// Set by goreleaser via -ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Load .env / .env.local from cwd upward so a project LINEAR_API_KEY is found.
	config.LoadDotEnv()

	app := &cli.Command{
		Name:            "linear",
		Usage:           "Linear CLI — issues, teams, and users from the terminal",
		Version:         fmt.Sprintf("%s (%s, %s)", version, commit, date),
		HideHelpCommand: true,
		Commands: []*cli.Command{
			commands.NewAuth(),
			commands.NewWhoami(),
			commands.NewIssue(),
			commands.NewTeam(),
			commands.NewUser(),
			commands.NewProject(),
			commands.NewMilestone(),
			commands.NewCycle(),
			commands.NewLabel(),
			commands.NewDocument(),
			commands.NewConfig(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
