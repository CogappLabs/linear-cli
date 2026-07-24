package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewStatus registers `linear status list` and `linear status view`.
func NewStatus() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Workflow status operations",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List workflow states for a team",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "team", Aliases: []string{"t"}, Required: true, Usage: "Team key"},
					&cli.BoolFlag{Name: "json"},
				},
				Action: func(_ context.Context, c *cli.Command) error {
					if err := requireAuth(); err != nil {
						return err
					}
					teamKey := c.String("team")
					if teamKey == "" {
						return fmt.Errorf("status list requires --team KEY")
					}
					states, err := client.ListStatuses(teamKey)
					if err != nil {
						return err
					}
					if c.Bool("json") {
						printJSON(states)
						return nil
					}
					for _, s := range states {
						label := format.ColorState(s.Name, s.Type)
						fmt.Printf("%s  (%s)\n", format.PadRight(label, 24), s.Type)
					}
					return nil
				},
			},
			{
				Name:      "view",
				Usage:     "View a single workflow state",
				ArgsUsage: "NAME",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "team", Aliases: []string{"t"}, Required: true, Usage: "Team key"},
					&cli.BoolFlag{Name: "json"},
				},
				Action: func(_ context.Context, c *cli.Command) error {
					if err := requireAuth(); err != nil {
						return err
					}
					teamKey := c.String("team")
					if teamKey == "" {
						return fmt.Errorf("status view requires --team KEY")
					}
					name := c.Args().First()
					if name == "" {
						return fmt.Errorf("status view requires NAME")
					}
					states, err := client.ListStatuses(teamKey)
					if err != nil {
						return err
					}
					var found *client.State
					for i := range states {
						if strings.EqualFold(states[i].Name, name) {
							found = &states[i]
							break
						}
					}
					if found == nil {
						return fmt.Errorf("status %q not found for team %s", name, teamKey)
					}
					if c.Bool("json") {
						printJSON(found)
						return nil
					}
					fmt.Printf("%s (%s)\n", format.ColorState(found.Name, found.Type), found.Type)
					return nil
				},
			},
		},
	}
}
