package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewMilestone registers `linear milestone …`.
func NewMilestone() *cli.Command {
	return &cli.Command{
		Name:  "milestone",
		Usage: "Work with project milestones",
		Commands: []*cli.Command{
			milestoneList(),
			milestoneView(),
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
