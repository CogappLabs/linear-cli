package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/config"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/parquet-go/parquet-go"
	"github.com/urfave/cli/v3"
)

// issueRow is the flat, parquet-friendly shape of an issue for bulk export.
type issueRow struct {
	Identifier  string `parquet:"identifier"`
	Title       string `parquet:"title"`
	State       string `parquet:"state"`
	StateType   string `parquet:"state_type"`
	Assignee    string `parquet:"assignee"`
	Team        string `parquet:"team"`
	Priority    string `parquet:"priority"`
	Labels      string `parquet:"labels"`
	URL         string `parquet:"url"`
	CreatedAt   string `parquet:"created_at"`
	UpdatedAt   string `parquet:"updated_at"`
	Description string `parquet:"description"`
}

func toRow(is client.Issue) issueRow {
	r := issueRow{
		Identifier:  is.Identifier,
		Title:       is.Title,
		Priority:    format.Priority(is.Priority),
		URL:         is.URL,
		CreatedAt:   is.CreatedAt,
		UpdatedAt:   is.UpdatedAt,
		Description: is.Description,
	}
	if is.State != nil {
		r.State = is.State.Name
		r.StateType = is.State.Type
	}
	if is.Assignee != nil {
		r.Assignee = is.Assignee.Name
	}
	if is.Team != nil {
		r.Team = is.Team.Key
	}
	for i, l := range is.Labels.Nodes {
		if i > 0 {
			r.Labels += ","
		}
		r.Labels += l.Name
	}
	return r
}

// issuePull bulk-fetches a team's issues, with an optional multi-status filter,
// and writes them to a Parquet file (or JSON / a table).
func issuePull() *cli.Command {
	return &cli.Command{
		Name:  "pull",
		Usage: "Bulk-fetch a team's issues (optionally by status) to Parquet, JSON, or a table",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "team", Aliases: []string{"t"}, Usage: "Team key (default: config defaultTeam)"},
			&cli.StringSliceFlag{Name: "state", Aliases: []string{"s"}, Usage: "State type (repeatable): backlog|unstarted|started|completed|canceled|triage|duplicate"},
			&cli.BoolFlag{Name: "not-done", Usage: "Exclude completed/canceled (overrides --state)"},
			&cli.StringFlag{Name: "parquet", Aliases: []string{"o"}, Usage: "Write results to this .parquet file"},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
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

			f := client.IssueFilter{
				TeamKey:    teamKey,
				StateTypes: c.StringSlice("state"),
				NotDone:    c.Bool("not-done"),
				All:        true,
			}
			issues, err := client.ListIssues(f)
			if err != nil {
				return err
			}

			if path := c.String("parquet"); path != "" {
				if err := writeIssueParquet(path, issues); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "Wrote %d issues to %s\n", len(issues), path)
				return nil
			}
			if c.Bool("json") {
				printJSON(issues)
				return nil
			}
			for _, is := range issues {
				renderIssueRow(is)
			}
			fmt.Fprintf(os.Stderr, "%d issues\n", len(issues))
			return nil
		},
	}
}

func writeIssueParquet(path string, issues []client.Issue) error {
	rows := make([]issueRow, len(issues))
	for i, is := range issues {
		rows[i] = toRow(is)
	}
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	w := parquet.NewGenericWriter[issueRow](fh)
	if _, err := w.Write(rows); err != nil {
		return err
	}
	return w.Close()
}
