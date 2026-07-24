package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/CogappLabs/linear-cli/internal/client"
	"github.com/CogappLabs/linear-cli/internal/config"
)

// jsonFlag is the shared --json output flag; append to per-command flag slices.
func requireAuth() error {
	key := config.ResolveAPIKey()
	if key == "" {
		return fmt.Errorf("no API key. Run `linear-cli auth login` or set LINEAR_API_KEY")
	}
	client.SetAPIKey(key)
	return nil
}

func printJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
