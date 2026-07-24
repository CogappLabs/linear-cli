package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewAgentSkill registers the `linear-cli agent-skill` command group.
func NewAgentSkill() *cli.Command {
	return &cli.Command{
		Name:  "agent-skill",
		Usage: "Work with agent skills",
		Commands: []*cli.Command{
			agentSkillList(),
			agentSkillView(),
		},
	}
}

func agentSkillList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List agent skills",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			skills, err := client.ListAgentSkills()
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(skills)
				return nil
			}
			for _, s := range skills {
				renderAgentSkillRow(s)
			}
			return nil
		},
	}
}

func renderAgentSkillRow(s client.AgentSkill) {
	fmt.Printf("%s  %s\n",
		format.PadRight(format.Cyan(s.Title), 30),
		s.Description,
	)
}

func agentSkillView() *cli.Command {
	return &cli.Command{
		Name:      "view",
		Aliases:   []string{"v"},
		Usage:     "Show an agent skill",
		ArgsUsage: "ID",
		Flags:     []cli.Flag{&cli.BoolFlag{Name: "json"}},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			id := c.Args().First()
			if id == "" {
				return fmt.Errorf("agent skill id required")
			}
			s, err := client.GetAgentSkill(id)
			if err != nil {
				return err
			}
			if s.ID == "" {
				return fmt.Errorf("agent skill %s not found", id)
			}
			if c.Bool("json") {
				printJSON(s)
				return nil
			}
			renderAgentSkill(s)
			return nil
		},
	}
}

func renderAgentSkill(s *client.AgentSkill) {
	fmt.Printf("%s\n", format.Bold(s.Title))
	if s.Creator != nil {
		fmt.Printf("Creator:  %s\n", s.Creator.Name)
	}
	if s.Owner != nil {
		fmt.Printf("Owner:    %s\n", s.Owner.Name)
	}
	fmt.Printf("Shared:   %v\n", s.Shared)
	fmt.Printf("URL:      %s\n", format.Gray(s.SlugID))
	if s.Description != "" {
		fmt.Printf("\n%s\n", s.Description)
	}
	if s.Body != "" {
		fmt.Printf("\n%s\n", s.Body)
	}
}
