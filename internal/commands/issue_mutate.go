package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/config"
	"github.com/urfave/cli/v3"
)

func issueCreate() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create an issue",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "title", Aliases: []string{"T"}, Usage: "Issue title (required)"},
			&cli.StringFlag{Name: "description", Aliases: []string{"d"}, Usage: "Markdown description"},
			&cli.StringFlag{Name: "team", Aliases: []string{"t"}, Usage: "Team key (default: config defaultTeam)"},
			&cli.StringFlag{Name: "assignee", Aliases: []string{"a"}, Usage: "Assignee email, or 'me'"},
			&cli.IntFlag{Name: "priority", Aliases: []string{"p"}, Usage: "0=none 1=urgent 2=high 3=medium 4=low"},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			title := c.String("title")
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			teamKey := c.String("team")
			if teamKey == "" {
				if cfg, _ := config.Load(); cfg != nil {
					teamKey = cfg.DefaultTeam
				}
			}
			if teamKey == "" {
				return fmt.Errorf("no team. Pass --team KEY or set a default team")
			}
			team, err := client.TeamByKey(teamKey)
			if err != nil {
				return err
			}

			input := map[string]any{
				"title":  title,
				"teamId": team.ID,
			}
			if d := c.String("description"); d != "" {
				input["description"] = d
			}
			if c.IsSet("priority") {
				input["priority"] = int(c.Int("priority"))
			}
			if a := c.String("assignee"); a != "" {
				uid, err := resolveUserID(a)
				if err != nil {
					return err
				}
				input["assigneeId"] = uid
			}

			is, err := client.CreateIssue(input)
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(is)
				return nil
			}
			fmt.Printf("Created %s: %s\n%s\n", is.Identifier, is.Title, is.URL)
			return nil
		},
	}
}

func issueUpdate() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update an issue",
		ArgsUsage: "[ID]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "title", Aliases: []string{"T"}},
			&cli.StringFlag{Name: "description", Aliases: []string{"d"}},
			&cli.StringFlag{Name: "state", Aliases: []string{"s"}, Usage: "Workflow state name (e.g. 'In Progress')"},
			&cli.StringFlag{Name: "assignee", Aliases: []string{"a"}, Usage: "Assignee email, or 'me'"},
			&cli.IntFlag{Name: "priority", Aliases: []string{"p"}},
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

			input := map[string]any{}
			if v := c.String("title"); v != "" {
				input["title"] = v
			}
			if v := c.String("description"); v != "" {
				input["description"] = v
			}
			if c.IsSet("priority") {
				input["priority"] = int(c.Int("priority"))
			}
			if a := c.String("assignee"); a != "" {
				uid, err := resolveUserID(a)
				if err != nil {
					return err
				}
				input["assigneeId"] = uid
			}
			if s := c.String("state"); s != "" {
				// Need the issue's team to resolve the state name to an id.
				cur, err := client.GetIssue(id)
				if err != nil {
					return err
				}
				if cur.Team == nil {
					return fmt.Errorf("could not determine team for %s", id)
				}
				sid, err := resolveStateID(cur.Team.ID, s)
				if err != nil {
					return err
				}
				input["stateId"] = sid
			}
			if len(input) == 0 {
				return fmt.Errorf("nothing to update. Pass at least one field flag")
			}

			is, err := client.UpdateIssue(id, input)
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(is)
				return nil
			}
			fmt.Printf("Updated %s\n", is.Identifier)
			return nil
		},
	}
}

func issueStart() *cli.Command {
	return &cli.Command{
		Name:      "start",
		Usage:     "Move an issue to a started state and assign it to you",
		ArgsUsage: "[ID]",
		Flags:     []cli.Flag{&cli.BoolFlag{Name: "json"}},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			id, err := resolveID(c)
			if err != nil {
				return err
			}
			cur, err := client.GetIssue(id)
			if err != nil {
				return err
			}
			if cur.Team == nil {
				return fmt.Errorf("could not determine team for %s", id)
			}
			sid, err := firstStateOfType(cur.Team.ID, "started")
			if err != nil {
				return err
			}
			me, err := client.Viewer()
			if err != nil {
				return err
			}
			is, err := client.UpdateIssue(id, map[string]any{
				"stateId":    sid,
				"assigneeId": me.ID,
			})
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(is)
				return nil
			}
			fmt.Printf("Started %s, assigned to you\n", is.Identifier)
			return nil
		},
	}
}

func issueComment() *cli.Command {
	return &cli.Command{
		Name:      "comment",
		Usage:     "Add a comment to an issue",
		ArgsUsage: "[ID] --body TEXT",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "body", Aliases: []string{"b"}, Usage: "Comment body (required)"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			id, err := resolveID(c)
			if err != nil {
				return err
			}
			body := c.String("body")
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			// commentCreate needs the issue UUID, not the identifier.
			is, err := client.GetIssue(id)
			if err != nil {
				return err
			}
			if is.ID == "" {
				return fmt.Errorf("issue %s not found", id)
			}
			if err := client.AddComment(is.ID, body); err != nil {
				return err
			}
			fmt.Printf("Commented on %s\n", is.Identifier)
			return nil
		},
	}
}

func issueCommentEdit() *cli.Command {
	return &cli.Command{
		Name:      "comment-edit",
		Usage:     "Edit a comment by its comment id",
		ArgsUsage: "COMMENT_ID --body TEXT",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "body", Aliases: []string{"b"}, Usage: "New comment body (required)"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			if c.Args().Len() != 1 {
				return fmt.Errorf("usage: linear-cli issue comment-edit COMMENT_ID --body TEXT")
			}
			body := c.String("body")
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			if err := client.EditComment(c.Args().First(), body); err != nil {
				return err
			}
			fmt.Println("Comment updated")
			return nil
		},
	}
}

func issueCommentDelete() *cli.Command {
	return &cli.Command{
		Name:      "comment-delete",
		Usage:     "Delete a comment by its comment id",
		ArgsUsage: "COMMENT_ID",
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			if c.Args().Len() != 1 {
				return fmt.Errorf("usage: linear-cli issue comment-delete COMMENT_ID")
			}
			if err := client.DeleteComment(c.Args().First()); err != nil {
				return err
			}
			fmt.Printf("Deleted comment %s\n", c.Args().First())
			return nil
		},
	}
}

// resolveUserID maps "me" or an email to a user UUID.
func resolveUserID(spec string) (string, error) {
	if spec == "me" {
		me, err := client.Viewer()
		if err != nil {
			return "", err
		}
		return me.ID, nil
	}
	users, err := client.ListUsers()
	if err != nil {
		return "", err
	}
	for _, u := range users {
		if strings.EqualFold(u.Email, spec) {
			return u.ID, nil
		}
	}
	return "", fmt.Errorf("no user with email %q", spec)
}

// resolveStateID maps a state name (case-insensitive) to its id within a team.
func resolveStateID(teamID, name string) (string, error) {
	states, err := client.TeamStates(teamID)
	if err != nil {
		return "", err
	}
	for _, s := range states {
		if strings.EqualFold(s.Name, name) {
			return s.ID, nil
		}
	}
	return "", fmt.Errorf("no state named %q in this team", name)
}

// firstStateOfType returns the id of the first state matching a type.
func firstStateOfType(teamID, stateType string) (string, error) {
	states, err := client.TeamStates(teamID)
	if err != nil {
		return "", err
	}
	for _, s := range states {
		if s.Type == stateType {
			return s.ID, nil
		}
	}
	return "", fmt.Errorf("no %q state in this team", stateType)
}
