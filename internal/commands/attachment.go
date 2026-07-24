package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewAttachment registers `linear attachment …`.
func NewAttachment() *cli.Command {
	return &cli.Command{
		Name:  "attachment",
		Usage: "Work with issue attachments",
		Commands: []*cli.Command{
			attachmentList(),
			attachmentAdd(),
			attachmentDelete(),
		},
	}
}

func attachmentList() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List attachments on an issue",
		ArgsUsage: "ISSUE",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			issue := c.Args().First()
			if issue == "" {
				return fmt.Errorf("ISSUE is required")
			}
			attachments, err := client.ListAttachments(issue)
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(attachments)
				return nil
			}
			for _, a := range attachments {
				renderAttachmentRow(a)
			}
			return nil
		},
	}
}

func renderAttachmentRow(a client.Attachment) {
	fmt.Println(format.PadRight(format.Cyan(a.Title), 40) + a.URL)
}

func attachmentAdd() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Attach a URL to an issue",
		ArgsUsage: "ISSUE URL",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "title", Usage: "Attachment title"},
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			if c.Args().Len() != 2 {
				return fmt.Errorf("ISSUE and URL are required")
			}
			issueArg := c.Args().Get(0)
			url := c.Args().Get(1)

			// attachmentLinkURL needs the issue UUID, not the identifier.
			issue, err := client.GetIssue(issueArg)
			if err != nil {
				return err
			}
			if issue.ID == "" {
				return fmt.Errorf("issue %s not found", issueArg)
			}

			attachment, err := client.AddAttachmentURL(issue.ID, url, c.String("title"))
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(attachment)
				return nil
			}
			fmt.Printf("Attached %s to %s\n", url, issue.Identifier)
			return nil
		},
	}
}

func attachmentDelete() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete an attachment",
		ArgsUsage: "ATTACHMENT_ID",
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			id := c.Args().First()
			if id == "" {
				return fmt.Errorf("ATTACHMENT_ID is required")
			}
			if err := client.DeleteAttachment(id); err != nil {
				return err
			}
			fmt.Printf("Deleted attachment %s\n", id)
			return nil
		},
	}
}
