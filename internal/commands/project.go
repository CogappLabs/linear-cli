package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewProject registers `linear project …`.
func NewProject() *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Work with projects",
		Commands: []*cli.Command{
			projectList(),
			projectView(),
			projectSave(),
		},
	}
}

func projectSave() *cli.Command {
	return &cli.Command{
		Name:      "save",
		Usage:     "Create a project (no ID) or update one (with ID)",
		ArgsUsage: "[ID]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Usage: "Project name (required to create)"},
			&cli.StringFlag{Name: "description", Aliases: []string{"d"}},
			&cli.StringFlag{Name: "team", Aliases: []string{"t"}, Usage: "Team key (required to create)"},
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
				p   *client.Project
				err error
			)
			if id == "" {
				// Create path.
				if _, ok := input["name"]; !ok {
					return fmt.Errorf("--name is required to create a project")
				}
				teamKey := c.String("team")
				if teamKey == "" {
					return fmt.Errorf("--team is required to create a project")
				}
				team, terr := client.TeamByKey(teamKey)
				if terr != nil {
					return terr
				}
				input["teamIds"] = []string{team.ID}
				p, err = client.CreateProject(input)
			} else {
				if len(input) == 0 {
					return fmt.Errorf("nothing to update — pass at least one field flag")
				}
				p, err = client.UpdateProject(id, input)
			}
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(p)
				return nil
			}
			fmt.Printf("Saved project %s\n%s\n", p.Name, p.URL)
			return nil
		},
	}
}

func projectList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List projects",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "team", Aliases: []string{"t"}, Usage: "Team key"},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			projects, err := client.ListProjects(c.String("team"))
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(projects)
				return nil
			}
			for _, p := range projects {
				renderProjectRow(p)
			}
			return nil
		},
	}
}

func renderProjectRow(p client.Project) {
	status := ""
	if p.Status != nil {
		status = p.Status.Name
	}
	fmt.Printf("%s  %s  %.0f%%\n",
		format.PadRight(format.Cyan(p.Name), 40),
		format.PadRight(status, 14),
		p.Progress*100,
	)
}

func projectView() *cli.Command {
	return &cli.Command{
		Name:      "view",
		Aliases:   []string{"v"},
		Usage:     "Show a project",
		ArgsUsage: "ID",
		Flags:     []cli.Flag{&cli.BoolFlag{Name: "json"}},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			id := c.Args().First()
			if id == "" {
				return fmt.Errorf("project id required")
			}
			p, err := client.GetProject(id)
			if err != nil {
				return err
			}
			if p.ID == "" {
				return fmt.Errorf("project %s not found", id)
			}
			if c.Bool("json") {
				printJSON(p)
				return nil
			}
			renderProject(p)
			return nil
		},
	}
}

func renderProject(p *client.Project) {
	fmt.Printf("%s\n", format.Bold(p.Name))
	if p.Status != nil {
		fmt.Printf("Status:   %s\n", p.Status.Name)
	}
	lead := "none"
	if p.Lead != nil {
		lead = p.Lead.Name
	}
	fmt.Printf("Lead:     %s\n", lead)
	fmt.Printf("Progress: %.0f%%\n", p.Progress*100)
	if p.StartDate != "" {
		fmt.Printf("Start:    %s\n", p.StartDate)
	}
	if p.TargetDate != "" {
		fmt.Printf("Target:   %s\n", p.TargetDate)
	}
	fmt.Printf("URL:      %s\n", format.Gray(p.URL))
	if p.Description != "" {
		fmt.Printf("\n%s\n", p.Description)
	}
}
