package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewDocument registers `linear document …`.
func NewDocument() *cli.Command {
	return &cli.Command{
		Name:  "document",
		Usage: "Work with documents",
		Commands: []*cli.Command{
			documentList(),
			documentView(),
			documentSave(),
		},
	}
}

func documentSave() *cli.Command {
	return &cli.Command{
		Name:      "save",
		Usage:     "Create a document (no ID) or update one (with ID)",
		ArgsUsage: "[ID]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "title", Aliases: []string{"T"}, Usage: "Document title (required to create)"},
			&cli.StringFlag{Name: "content", Aliases: []string{"c"}, Usage: "Markdown content"},
			&cli.StringFlag{Name: "team", Aliases: []string{"t"}, Usage: "Team key (create: team or project required)"},
			&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "Project UUID (create: team or project required)"},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			input := map[string]any{}
			if v := c.String("title"); v != "" {
				input["title"] = v
			}
			if v := c.String("content"); v != "" {
				input["content"] = v
			}

			id := c.Args().First()
			var (
				d   *client.Document
				err error
			)
			if id == "" {
				if _, ok := input["title"]; !ok {
					return fmt.Errorf("--title is required to create a document")
				}
				team := c.String("team")
				proj := c.String("project")
				if team == "" && proj == "" {
					return fmt.Errorf("create needs --team KEY or --project UUID")
				}
				if team != "" {
					t, terr := client.TeamByKey(team)
					if terr != nil {
						return terr
					}
					input["teamId"] = t.ID
				}
				if proj != "" {
					input["projectId"] = proj
				}
				d, err = client.CreateDocument(input)
			} else {
				if len(input) == 0 {
					return fmt.Errorf("nothing to update — pass at least one field flag")
				}
				d, err = client.UpdateDocument(id, input)
			}
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(d)
				return nil
			}
			fmt.Printf("Saved document %s\n%s\n", d.Title, d.URL)
			return nil
		},
	}
}

func documentList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List documents",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			documents, err := client.ListDocuments()
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(documents)
				return nil
			}
			for _, d := range documents {
				renderDocumentRow(d)
			}
			return nil
		},
	}
}

func renderDocumentRow(d client.Document) {
	creator := ""
	if d.Creator != nil {
		creator = d.Creator.Name
	}
	fmt.Printf("%s  %s  %s\n",
		format.PadRight(format.Cyan(d.Title), 50),
		format.PadRight(creator, 20),
		format.TimeAgo(d.UpdatedAt),
	)
}

func documentView() *cli.Command {
	return &cli.Command{
		Name:      "view",
		Aliases:   []string{"v"},
		Usage:     "Show a document",
		ArgsUsage: "ID",
		Flags:     []cli.Flag{&cli.BoolFlag{Name: "json"}},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			id := c.Args().First()
			if id == "" {
				return fmt.Errorf("document id required")
			}
			d, err := client.GetDocument(id)
			if err != nil {
				return err
			}
			if d.ID == "" {
				return fmt.Errorf("document %s not found", id)
			}
			if c.Bool("json") {
				printJSON(d)
				return nil
			}
			renderDocument(d)
			return nil
		},
	}
}

func renderDocument(d *client.Document) {
	fmt.Printf("%s\n", format.Bold(d.Title))
	if d.Creator != nil {
		fmt.Printf("Creator:  %s\n", d.Creator.Name)
	}
	if d.Project != nil {
		fmt.Printf("Project:  %s\n", d.Project.Name)
	}
	fmt.Printf("URL:      %s\n", format.Gray(d.URL))
	if d.Content != "" {
		fmt.Printf("\n%s\n", d.Content)
	}
}
