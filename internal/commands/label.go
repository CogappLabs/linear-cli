package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewLabel registers the `linear-cli label` command group.
func NewLabel() *cli.Command {
	return &cli.Command{
		Name:  "label",
		Usage: "Work with issue labels",
		Commands: []*cli.Command{
			labelList(),
			labelCreate(),
		},
	}
}

func labelList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List issue labels",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "team", Aliases: []string{"t"}, Usage: "Team key"},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			labels, err := client.ListLabels(c.String("team"))
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(labels)
				return nil
			}
			for _, l := range labels {
				renderLabelRow(l)
			}
			return nil
		},
	}
}

func renderLabelRow(l client.LabelDetail) {
	line := format.PadRight(format.Cyan(l.Name), 30) + l.Color
	if l.Team != nil {
		line += "  " + l.Team.Key
	}
	fmt.Println(line)
}

func labelCreate() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create an issue label",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "team", Aliases: []string{"t"}, Usage: "Team key", Required: true},
			&cli.StringFlag{Name: "name", Required: true},
			&cli.StringFlag{Name: "color"},
			&cli.StringFlag{Name: "description"},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			team, err := client.TeamByKey(c.String("team"))
			if err != nil {
				return err
			}
			input := map[string]any{
				"teamId": team.ID,
				"name":   c.String("name"),
			}
			if v := c.String("color"); v != "" {
				input["color"] = v
			}
			if v := c.String("description"); v != "" {
				input["description"] = v
			}
			label, err := client.CreateLabel(input)
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(label)
				return nil
			}
			fmt.Printf("Created label %s\n", label.Name)
			return nil
		},
	}
}
