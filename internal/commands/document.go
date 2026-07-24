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
