package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewRelease registers the `linear-cli release` command group.
func NewRelease() *cli.Command {
	return &cli.Command{
		Name:  "release",
		Usage: "Work with releases",
		Commands: []*cli.Command{
			releaseList(),
			releaseView(),
			releaseNotes(),
		},
	}
}

func releaseList() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List releases",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			releases, err := client.ListReleases()
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(releases)
				return nil
			}
			for _, r := range releases {
				renderReleaseRow(r)
			}
			return nil
		},
	}
}

func renderReleaseRow(r client.Release) {
	stage := ""
	if r.Stage != nil {
		stage = r.Stage.Name
	}
	fmt.Printf("%s  %s  %s\n",
		format.PadRight(format.Cyan(r.Name), 40),
		format.PadRight(stage, 15),
		format.TimeAgo(r.UpdatedAt),
	)
}

func releaseView() *cli.Command {
	return &cli.Command{
		Name:      "view",
		Aliases:   []string{"v"},
		Usage:     "Show a release",
		ArgsUsage: "ID",
		Flags:     []cli.Flag{&cli.BoolFlag{Name: "json"}},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			id := c.Args().First()
			if id == "" {
				return fmt.Errorf("release id required")
			}
			r, err := client.GetRelease(id)
			if err != nil {
				return err
			}
			if r.ID == "" {
				return fmt.Errorf("release %s not found", id)
			}
			if c.Bool("json") {
				printJSON(r)
				return nil
			}
			renderRelease(r)
			return nil
		},
	}
}

func renderRelease(r *client.Release) {
	fmt.Printf("%s\n", format.Bold(r.Name))
	if r.Version != "" {
		fmt.Printf("Version:  %s\n", r.Version)
	}
	if r.Stage != nil {
		fmt.Printf("Stage:    %s\n", r.Stage.Name)
	}
	if r.Pipeline != nil {
		fmt.Printf("Pipeline: %s\n", r.Pipeline.Name)
	}
	if r.Creator != nil {
		fmt.Printf("Creator:  %s\n", r.Creator.Name)
	}
	fmt.Printf("URL:      %s\n", format.Gray(r.URL))
	if r.Description != "" {
		fmt.Printf("\n%s\n", r.Description)
	}
}

func releaseNotes() *cli.Command {
	return &cli.Command{
		Name:      "notes",
		Usage:     "List release notes covering a release",
		ArgsUsage: "RELEASE_ID",
		Flags:     []cli.Flag{&cli.BoolFlag{Name: "json"}},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			releaseID := c.Args().First()
			if releaseID == "" {
				return fmt.Errorf("release id required")
			}
			notes, err := client.ListReleaseNotes(releaseID)
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(notes)
				return nil
			}
			for _, n := range notes {
				renderReleaseNoteRow(n)
			}
			return nil
		},
	}
}

func renderReleaseNoteRow(n client.ReleaseNote) {
	title := n.Title
	if title == "" {
		title = n.SlugID
	}
	fmt.Printf("%s  %s\n",
		format.PadRight(format.Cyan(title), 50),
		format.TimeAgo(n.UpdatedAt),
	)
}
