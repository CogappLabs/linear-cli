package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

func issueRelate() *cli.Command {
	return &cli.Command{
		Name:      "relate",
		Usage:     "Link two issues (ID is the subject: ID blocks OTHER-ID, etc.)",
		ArgsUsage: "ID OTHER-ID",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "type", Value: "related", Usage: "related|blocks|blocked-by|duplicate (duplicate: ID is a duplicate of OTHER-ID)"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			if c.Args().Len() != 2 {
				return fmt.Errorf("usage: issue relate ID OTHER-ID --type related|blocks|blocked-by|duplicate")
			}
			relType := strings.ToLower(c.String("type"))
			swap := relType == "blocked-by"
			switch relType {
			case "related", "blocks", "duplicate":
			case "blocked-by":
				relType = "blocks"
			default:
				return fmt.Errorf("unknown --type %q (want related|blocks|blocked-by|duplicate)", c.String("type"))
			}
			a, b, err := resolveIssuePair(c)
			if err != nil {
				return err
			}
			if swap {
				a, b = b, a
			}
			if err := client.CreateRelation(a.ID, b.ID, relType); err != nil {
				return err
			}
			fmt.Printf("%s %s %s\n", a.Identifier, relType, b.Identifier)
			return nil
		},
	}
}

func issueUnrelate() *cli.Command {
	return &cli.Command{
		Name:      "unrelate",
		Usage:     "Remove any relation between two issues (either direction)",
		ArgsUsage: "ID OTHER-ID",
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			if c.Args().Len() != 2 {
				return fmt.Errorf("usage: issue unrelate ID OTHER-ID")
			}
			a, b, err := resolveIssuePair(c)
			if err != nil {
				return err
			}
			ids := client.RelationIDsBetween(a, b.Identifier)
			if len(ids) == 0 {
				return fmt.Errorf("no relation between %s and %s", a.Identifier, b.Identifier)
			}
			for _, id := range ids {
				if err := client.DeleteRelation(id); err != nil {
					return err
				}
			}
			fmt.Printf("Removed %d relation(s) between %s and %s\n", len(ids), a.Identifier, b.Identifier)
			return nil
		},
	}
}

// resolveIssuePair fetches both issues named by the two args, which also
// resolves identifiers to UUIDs.
func resolveIssuePair(c *cli.Command) (*client.Issue, *client.Issue, error) {
	var out [2]*client.Issue
	for i := range out {
		id := strings.ToUpper(c.Args().Get(i))
		is, err := client.GetIssue(id)
		if err != nil {
			return nil, nil, err
		}
		if is.ID == "" {
			return nil, nil, fmt.Errorf("issue %s not found", id)
		}
		out[i] = is
	}
	return out[0], out[1], nil
}

// renderRelations prints relations from this issue's point of view. A row in
// Relations reads "this <type> other"; a row in InverseRelations reads
// "other <type> this".
func renderRelations(is *client.Issue) {
	groups := map[string][]string{}
	var order []string
	add := func(label string, other *client.RelIssue) {
		if other == nil {
			return
		}
		if _, ok := groups[label]; !ok {
			order = append(order, label)
		}
		groups[label] = append(groups[label], format.Cyan(other.Identifier)+" "+other.Title)
	}
	for _, r := range is.Relations.Nodes {
		switch r.Type {
		case "blocks":
			add("Blocks", r.RelatedIssue)
		case "duplicate":
			add("Duplicate of", r.RelatedIssue)
		case "related":
			add("Related", r.RelatedIssue)
		default:
			add(strings.ToUpper(r.Type[:1])+r.Type[1:], r.RelatedIssue)
		}
	}
	for _, r := range is.InverseRelations.Nodes {
		switch r.Type {
		case "blocks":
			add("Blocked by", r.Issue)
		case "duplicate":
			add("Duplicated by", r.Issue)
		case "related":
			add("Related", r.Issue)
		default:
			add(strings.ToUpper(r.Type[:1])+r.Type[1:], r.Issue)
		}
	}
	for _, label := range order {
		for _, line := range groups[label] {
			fmt.Printf("%s %s\n", format.PadRight(label+":", 14), line)
		}
	}
}
