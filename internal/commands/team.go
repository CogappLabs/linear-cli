package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewTeam registers the `linear-cli team` command group.
func NewTeam() *cli.Command {
	return &cli.Command{
		Name:  "team",
		Usage: "Team operations",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List teams",
				Flags: []cli.Flag{&cli.BoolFlag{Name: "json"}},
				Action: func(_ context.Context, c *cli.Command) error {
					if err := requireAuth(); err != nil {
						return err
					}
					teams, err := client.ListTeams()
					if err != nil {
						return err
					}
					if c.Bool("json") {
						printJSON(teams)
						return nil
					}
					for _, t := range teams {
						fmt.Printf("%s  %s\n", format.PadRight(format.Cyan(t.Key), 16), t.Name)
					}
					return nil
				},
			},
			{
				Name:      "view",
				Usage:     "Show a team by key",
				ArgsUsage: "KEY",
				Flags:     []cli.Flag{&cli.BoolFlag{Name: "json"}},
				Action: func(_ context.Context, c *cli.Command) error {
					if err := requireAuth(); err != nil {
						return err
					}
					if c.Args().Len() != 1 {
						return fmt.Errorf("usage: linear-cli team view KEY")
					}
					t, err := client.GetTeam(c.Args().First())
					if err != nil {
						return err
					}
					if c.Bool("json") {
						printJSON(t)
						return nil
					}
					fmt.Printf("%s  %s\n", format.Cyan(t.Key), t.Name)
					return nil
				},
			},
		},
	}
}
