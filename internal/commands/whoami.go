package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/urfave/cli/v3"
)

// NewWhoami registers `linear whoami`.
func NewWhoami() *cli.Command {
	return &cli.Command{
		Name:  "whoami",
		Usage: "Show the authenticated user",
		Flags: []cli.Flag{&cli.BoolFlag{Name: "json"}},
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireAuth(); err != nil {
				return err
			}
			u, err := client.Viewer()
			if err != nil {
				return err
			}
			if c.Bool("json") {
				printJSON(u)
				return nil
			}
			fmt.Printf("%s <%s>\n", u.Name, u.Email)
			return nil
		},
	}
}
