package commands

import (
	"context"
	"fmt"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/format"
	"github.com/urfave/cli/v3"
)

// NewUser registers `linear user list`.
func NewUser() *cli.Command {
	return &cli.Command{
		Name:  "user",
		Usage: "User operations",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List workspace members",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "json"},
					&cli.BoolFlag{Name: "all", Usage: "Include deactivated users"},
				},
				Action: func(_ context.Context, c *cli.Command) error {
					if err := requireAuth(); err != nil {
						return err
					}
					users, err := client.ListUsers()
					if err != nil {
						return err
					}
					if c.Bool("json") {
						printJSON(users)
						return nil
					}
					for _, u := range users {
						if !u.Active && !c.Bool("all") {
							continue
						}
						name := u.Name
						if !u.Active {
							name = format.Gray(name + " (inactive)")
						}
						fmt.Printf("%s  %s\n", format.PadRight(name, 28), format.Gray(u.Email))
					}
					return nil
				},
			},
			{
				Name:      "view",
				Usage:     "Show a user by email",
				ArgsUsage: "EMAIL",
				Flags:     []cli.Flag{&cli.BoolFlag{Name: "json"}},
				Action: func(_ context.Context, c *cli.Command) error {
					if err := requireAuth(); err != nil {
						return err
					}
					if c.Args().Len() != 1 {
						return fmt.Errorf("usage: linear user view EMAIL")
					}
					u, err := client.GetUser(c.Args().First())
					if err != nil {
						return err
					}
					if c.Bool("json") {
						printJSON(u)
						return nil
					}
					state := "active"
					if !u.Active {
						state = "inactive"
					}
					fmt.Printf("%s <%s>  [%s]\n", u.Name, u.Email, state)
					return nil
				},
			},
		},
	}
}
