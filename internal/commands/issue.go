package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/config"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/CogappLabs/linear-cli/internal/gitutil"
	"github.com/urfave/cli/v3"
)

// NewIssue registers `linear issue …`.
func NewIssue() *cli.Command {
	return &cli.Command{
		Name:    "issue",
		Aliases: []string{"i"},
		Usage:   "Work with issues",
		Commands: []*cli.Command{
			issueList(),
			issuePull(),
			issueView(),
			issueCreate(),
			issueUpdate(),
			issueStart(),
			issueComment(),
			issueComments(),
			issueCommentEdit(),
			issueCommentDelete(),
			issueURL(),
			issueBranch(),
		},
	}
}

// resolveID returns the first arg, or falls back to the issue id parsed from the
// current git branch. Errors if neither is available.
func resolveID(c *cli.Command) (string, error) {
	if c.Args().Len() > 0 {
		return strings.ToUpper(c.Args().First()), nil
	}
	if id := gitutil.IssueIDFromBranch(); id != "" {
		return id, nil
	}
	return "", fmt.Errorf("no issue id given and none found in the current git branch")
}

func issueList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List issues",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "team", Aliases: []string{"t"}, Usage: "Team key (default: config defaultTeam)"},
			&cli.BoolFlag{Name: "mine", Aliases: []string{"m"}, Usage: "Only issues assigned to me"},
			&cli.StringSliceFlag{Name: "state", Aliases: []string{"s"}, Usage: "State type (repeatable): backlog|unstarted|started|completed|canceled|triage|duplicate"},
			&cli.BoolFlag{Name: "not-done", Usage: "Exclude completed/canceled (overrides --state)"},
			&cli.IntFlag{Name: "limit", Aliases: []string{"n"}, Value: 50, Usage: "Max results (ignored with --all)"},
			&cli.BoolFlag{Name: "all", Usage: "Fetch every match, paging through all results"},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			f := client.IssueFilter{
				TeamKey:    c.String("team"),
				StateTypes: c.StringSlice("state"),
				NotDone:    c.Bool("not-done"),
				First:      int(c.Int("limit")),
				All:        c.Bool("all"),
			}
			if f.TeamKey == "" {
				if cfg, _ := config.Load(); cfg != nil {
					f.TeamKey = cfg.DefaultTeam
				}
			}
			if c.Bool("mine") {
				me, err := client.Viewer()
				if err != nil {
					return err
				}
				f.AssigneeID = me.ID
			}
			issues, err := client.ListIssues(f)
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(issues)
				return nil
			}
			for _, is := range issues {
				renderIssueRow(is)
			}
			return nil
		},
	}
}

func renderIssueRow(is client.Issue) {
	state := ""
	if is.State != nil {
		state = format.ColorState(is.State.Name, is.State.Type)
	}
	assignee := "unassigned"
	if is.Assignee != nil {
		assignee = is.Assignee.Name
	}
	fmt.Printf("%s  %s  %s  %s\n",
		format.PadRight(format.Cyan(is.Identifier), 18),
		format.PadRight(state, 24),
		format.PadRight(assignee, 22),
		is.Title,
	)
}

func issueView() *cli.Command {
	return &cli.Command{
		Name:      "view",
		Aliases:   []string{"v"},
		Usage:     "Show an issue (defaults to the current branch's issue)",
		ArgsUsage: "[ID]",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json"},
			&cli.BoolFlag{Name: "comments", Aliases: []string{"c"}, Usage: "Include comments"},
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
				printJSON(is)
				return nil
			}
			renderIssue(is, c.Bool("comments"))
			return nil
		},
	}
}

func renderIssue(is *client.Issue, withComments bool) {
	fmt.Printf("%s  %s\n", format.Cyan(is.Identifier), format.Bold(is.Title))
	if is.State != nil {
		fmt.Printf("State:    %s\n", format.ColorState(is.State.Name, is.State.Type))
	}
	assignee := "unassigned"
	if is.Assignee != nil {
		assignee = is.Assignee.Name
	}
	fmt.Printf("Assignee: %s\n", assignee)
	fmt.Printf("Priority: %s\n", format.Priority(is.Priority))
	if len(is.Labels.Nodes) > 0 {
		names := make([]string, len(is.Labels.Nodes))
		for i, l := range is.Labels.Nodes {
			names[i] = l.Name
		}
		fmt.Printf("Labels:   %s\n", strings.Join(names, ", "))
	}
	fmt.Printf("URL:      %s\n", format.Gray(is.URL))
	if is.Description != "" {
		fmt.Printf("\n%s\n", is.Description)
	}
	if withComments && len(is.Comments.Nodes) > 0 {
		fmt.Printf("\n%s\n", format.Bold("Comments"))
		for _, cm := range is.Comments.Nodes {
			who := ""
			if cm.User != nil {
				who = cm.User.Name
			}
			fmt.Printf("  %s (%s):\n  %s\n", format.Cyan(who), format.Gray(format.TimeAgo(cm.CreatedAt)), cm.Body)
		}
	}
}

func issueURL() *cli.Command {
	return &cli.Command{
		Name:      "url",
		Usage:     "Print an issue's URL",
		ArgsUsage: "[ID]",
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
			fmt.Println(is.URL)
			return nil
		},
	}
}

func issueBranch() *cli.Command {
	return &cli.Command{
		Name:      "branch",
		Usage:     "Print the suggested git branch name for an issue",
		ArgsUsage: "[ID]",
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
			fmt.Println(is.BranchName)
			return nil
		},
	}
}
