package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewMilestone registers the `linear-cli milestone` command group.
func NewMilestone() *cli.Command {
	return &cli.Command{
		Name:  "milestone",
		Usage: "Work with project milestones",
		Commands: []*cli.Command{
			milestoneList(),
			milestoneView(),
			milestoneSave(),
		},
	}
}

func milestoneSave() *cli.Command {
	return &cli.Command{
		Name:      "save",
		Usage:     "Create a milestone (no ID) or update one (with ID)",
		ArgsUsage: "[ID]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Usage: "Milestone name (required to create)"},
			&cli.StringFlag{Name: "description", Aliases: []string{"d"}},
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "Project UUID (required to create)"},
			&cli.StringFlag{Name: "target-date", Usage: "Target date (YYYY-MM-DD)"},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			input := map[string]any{}
			if v := c.String("name"); v != "" {
				input["name"] = v
			}
			if v := c.String("description"); v != "" {
				input["description"] = v
			}
			if v := c.String("target-date"); v != "" {
				input["targetDate"] = v
			}

			id := c.Args().First()
			var (
				m   *client.Milestone
				err error
			)
			if id == "" {
				if _, ok := input["name"]; !ok {
					return fmt.Errorf("--name is required to create a milestone")
				}
				proj := c.String("project")
				if proj == "" {
					return fmt.Errorf("--project is required to create a milestone")
				}
				input["projectId"] = proj
				m, err = client.CreateMilestone(input)
			} else {
				if len(input) == 0 {
					return fmt.Errorf("nothing to update. Pass at least one field flag")
				}
				m, err = client.UpdateMilestone(id, input)
			}
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(m)
				return nil
			}
			fmt.Printf("Saved milestone %s\n", m.Name)
			return nil
		},
	}
}

func milestoneList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List milestones for a project",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "Project UUID", Required: true},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			milestones, err := client.ListMilestones(c.String("project"))
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(milestones)
				return nil
			}
			for _, m := range milestones {
				renderMilestoneRow(m)
			}
			return nil
		},
	}
}

func renderMilestoneRow(m client.Milestone) {
	fmt.Printf("%s  %.0f%%  %s\n",
		format.PadRight(format.Cyan(m.Name), 40),
		m.Progress*100,
		m.TargetDate,
	)
}

func milestoneView() *cli.Command {
	return &cli.Command{
		Name:      "view",
		Aliases:   []string{"v"},
		Usage:     "Show a milestone",
		ArgsUsage: "ID",
		Flags:     []cli.Flag{&cli.BoolFlag{Name: "json"}},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			id := c.Args().First()
			if id == "" {
				return fmt.Errorf("milestone id required")
			}
			m, err := client.GetMilestone(id)
			if err != nil {
				return err
			}
			if m.ID == "" {
				return fmt.Errorf("milestone %s not found", id)
			}
			if c.Bool("json") {
				printJSON(m)
				return nil
			}
			renderMilestone(m)
			return nil
		},
	}
}

func renderMilestone(m *client.Milestone) {
	fmt.Printf("%s\n", format.Bold(m.Name))
	if m.Project != nil {
		fmt.Printf("Project:  %s\n", m.Project.Name)
	}
	fmt.Printf("Progress: %.0f%%\n", m.Progress*100)
	if m.TargetDate != "" {
		fmt.Printf("Target:   %s\n", m.TargetDate)
	}
	if m.Description != "" {
		fmt.Printf("\n%s\n", m.Description)
	}
}
