package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/config"
	"github.com/urfave/cli/v3"
)

// NewConfig registers `linear config set-team|show`.
func NewConfig() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "View or change CLI configuration",
		Commands: []*cli.Command{
			{
				Name:      "set-team",
				Usage:     "Set the default team key used when --team is omitted",
				ArgsUsage: "KEY",
				Action: func(_ context.Context, c *cli.Command) error {
					if c.Args().Len() != 1 {
						return fmt.Errorf("usage: linear config set-team KEY")
					}
					cfg, err := config.Load()
					if err != nil {
						return err
					}
					cfg.DefaultTeam = c.Args().First()
					if err := config.Save(cfg); err != nil {
						return err
					}
					fmt.Printf("Default team set to %s\n", cfg.DefaultTeam)
					return nil
				},
			},
			{
				Name:  "show",
				Usage: "Show current configuration",
				Action: func(_ context.Context, _ *cli.Command) error {
					cfg, err := config.Load()
					if err != nil {
						return err
					}
					keyState := "not set"
					if cfg.APIKey != "" {
						keyState = "set (config file)"
					}
					fmt.Printf("API key:      %s\n", keyState)
					fmt.Printf("Default team: %s\n", orDash(cfg.DefaultTeam))
					return nil
				},
			},
		},
	}
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
