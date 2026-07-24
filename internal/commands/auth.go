package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/config"
	"github.com/urfave/cli/v3"
)

// NewAuth registers `linear auth login|status`.
func NewAuth() *cli.Command {
	return &cli.Command{
		Name:  "auth",
		Usage: "Manage the Linear API key",
		Commands: []*cli.Command{
			{
				Name:  "login",
				Usage: "Store a Linear personal API key",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "key", Usage: "API key (else read from stdin prompt)"},
				},
				Action: func(_ context.Context, c *cli.Command) error {
					key := c.String("key")
					if key == "" {
						fmt.Fprint(os.Stderr, "Paste your Linear API key (from linear.app/settings/account/security): ")
						r := bufio.NewReader(os.Stdin)
						line, _ := r.ReadString('\n')
						key = strings.TrimSpace(line)
					}
					if key == "" {
						return fmt.Errorf("no key provided")
					}
					// Verify before saving.
					client.SetAPIKey(key)
					u, err := client.Viewer()
					if err != nil {
						return fmt.Errorf("key rejected: %w", err)
					}
					cfg, err := config.Load()
					if err != nil {
						return err
					}
					cfg.APIKey = key
					if err := config.Save(cfg); err != nil {
						return err
					}
					fmt.Printf("Logged in as %s (%s)\n", u.Name, u.Email)
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Show the current authenticated user",
				Action: func(_ context.Context, _ *cli.Command) error {
					if err := requireAuth(); err != nil {
						return err
					}
					u, err := client.Viewer()
					if err != nil {
						return err
					}
					fmt.Printf("Logged in as %s (%s)\n", u.Name, u.Email)
					return nil
				},
			},
		},
	}
}
