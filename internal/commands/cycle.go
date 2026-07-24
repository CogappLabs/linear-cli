package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewCycle registers `linear cycle list`.
func NewCycle() *cli.Command {
	return &cli.Command{
		Name:  "cycle",
		Usage: "Cycle operations",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List cycles for a team",
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
						return fmt.Errorf("cycle list requires --team KEY")
					}
					cycles, err := client.ListCycles(teamKey)
					if err != nil {
						return err
					}
					if c.Bool("json") {
						printJSON(cycles)
						return nil
					}
					for _, cy := range cycles {
						label := fmt.Sprintf("Cycle %d", int(cy.Number))
						window := fmt.Sprintf("%s → %s", format.TimeAgo(cy.StartsAt), format.TimeAgo(cy.EndsAt))
						fmt.Printf("%s  %s  %s\n", format.PadRight(format.Cyan(label), 12), format.PadRight(cy.Name, 24), window)
					}
					return nil
				},
			},
		},
	}
}
