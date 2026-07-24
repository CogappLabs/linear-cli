package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

func issueComments() *cli.Command {
	return &cli.Command{
		Name:      "comments",
		Aliases:   []string{"c"},
		Usage:     "List comments on an issue",
		ArgsUsage: "[ID]",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			id, err := resolveID(c)
			if err != nil {
				return err
			}
			is, err := client.GetIssue(id)
			if err != nil {
				return err
			}
			if is.ID == "" {
				return fmt.Errorf("issue %s not found", id)
			}
			if c.Bool("json") {
				printJSON(is.Comments.Nodes)
				return nil
			}
			for _, cm := range is.Comments.Nodes {
				who := ""
				if cm.User != nil {
					who = cm.User.Name
				}
				fmt.Printf("%s (%s):\n  %s\n", format.Cyan(who), format.Gray(format.TimeAgo(cm.CreatedAt)), cm.Body)
			}
			return nil
		},
	}
}
