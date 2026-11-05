package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/urfave/cli/v3"
)

type exportComment struct {
	Author    string `json:"author"`
	CreatedAt string `json:"createdAt"`
	Body      string `json:"body"`
}

type exportMedia struct {
	Name string `json:"name"`
	Path string `json:"path"` // relative to the export directory
}

type exportIssue struct {
	Identifier  string          `json:"identifier"`
	Title       string          `json:"title"`
	State       string          `json:"state"`
	Assignee    string          `json:"assignee"`
	Priority    int             `json:"priority"`
	Labels      []string        `json:"labels"`
	URL         string          `json:"url"`
	Description string          `json:"description"`
	Comments    []exportComment `json:"comments"`
	Media       []exportMedia   `json:"media"`
}

var unsafeFilename = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func issueExport() *cli.Command {
	return &cli.Command{
		Name:      "export",
		Usage:     "Export issues with comments and download their uploaded images and videos",
		ArgsUsage: "ID [ID...]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "out", Aliases: []string{"o"}, Value: "linear-export", Usage: "Output directory"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			if c.Args().Len() == 0 {
				return fmt.Errorf("give at least one issue id")
			}
			dir := c.String("out")

			var issues []exportIssue
			for _, arg := range c.Args().Slice() {
				id := strings.ToUpper(arg)
				fmt.Fprintf(os.Stderr, "Exporting %s…\n", id)
				is, err := client.GetIssue(id)
				if err != nil {
					return err
				}
				comments, err := client.ListComments(id)
				if err != nil {
					return err
				}

				ex := exportIssue{
					Identifier:  is.Identifier,
					Title:       is.Title,
					URL:         is.URL,
					Priority:    is.Priority,
					Description: is.Description,
					Labels:      []string{},
					Comments:    []exportComment{},
					Media:       []exportMedia{},
				}
				if is.State != nil {
					ex.State = is.State.Name
				}
				if is.Assignee != nil {
					ex.Assignee = is.Assignee.Name
				}
				for _, l := range is.Labels.Nodes {
					ex.Labels = append(ex.Labels, l.Name)
				}

				uploads := client.FindUploads(is.Description)
				for _, cm := range comments {
					author := ""
					if cm.User != nil {
						author = cm.User.Name
					}
					ex.Comments = append(ex.Comments, exportComment{Author: author, CreatedAt: cm.CreatedAt, Body: cm.Body})
					uploads = append(uploads, client.FindUploads(cm.Body)...)
				}

				seen := map[string]bool{}
				for _, up := range uploads {
					if seen[up.URL] {
						continue
					}
					seen[up.URL] = true
					name := unsafeFilename.ReplaceAllString(up.Name, "-")
					if name == "" || name == "-" {
						name = "file"
					}
					rel := filepath.Join("media", is.Identifier, fmt.Sprintf("%d-%s", len(ex.Media)+1, name))
					if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(rel)), 0o755); err != nil {
						return err
					}
					fmt.Fprintf(os.Stderr, "  downloading %s\n", up.Name)
					if err := client.DownloadUpload(up.URL, filepath.Join(dir, rel)); err != nil {
						return err
					}
					ex.Media = append(ex.Media, exportMedia{Name: up.Name, Path: rel})
				}
				issues = append(issues, ex)
			}

			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(issues, "", "  ")
			out := filepath.Join(dir, "issues.json")
			if err := os.WriteFile(out, b, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Wrote %s (%d issues)\n", out, len(issues))
			return nil
		},
	}
}
