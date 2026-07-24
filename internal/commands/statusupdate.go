package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewStatusUpdate registers `linear status-update …`.
func NewStatusUpdate() *cli.Command {
	return &cli.Command{
		Name:  "status-update",
		Usage: "Work with project status updates",
		Commands: []*cli.Command{
			statusUpdateList(),
			statusUpdateCreate(),
			statusUpdateDelete(),
		},
	}
}

func statusUpdateList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List status updates for a project",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "Project UUID", Required: true},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			updates, err := client.ListStatusUpdates(c.String("project"))
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(updates)
				return nil
			}
			for _, u := range updates {
				renderStatusUpdateRow(u)
			}
			return nil
		},
	}
}

func renderStatusUpdateRow(u client.StatusUpdate) {
	line := firstLine(u.Body)
	fmt.Printf("%s  %s  %s\n",
		format.Gray(format.TimeAgo(u.CreatedAt)),
		u.Health,
		line,
	)
}

func firstLine(body string) string {
	if i := strings.IndexByte(body, '\n'); i != -1 {
		return body[:i]
	}
	return body
}

func statusUpdateCreate() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Post a status update to a project",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "Project UUID", Required: true},
			&cli.StringFlag{Name: "body", Aliases: []string{"b"}, Required: true},
			&cli.StringFlag{Name: "health", Usage: "onTrack, atRisk, or offTrack"},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			input := map[string]any{
				"projectId": c.String("project"),
				"body":      c.String("body"),
			}
			if v := c.String("health"); v != "" {
				input["health"] = v
			}
			update, err := client.CreateStatusUpdate(input)
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(update)
				return nil
			}
			fmt.Println("Posted status update")
			return nil
		},
	}
}

func statusUpdateDelete() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a status update",
		ArgsUsage: "ID",
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			id := c.Args().First()
			if id == "" {
				return fmt.Errorf("status update id required")
			}
			if err := client.DeleteStatusUpdate(id); err != nil {
				return err
			}
			fmt.Printf("Deleted status update %s\n", id)
			return nil
		},
	}
}
