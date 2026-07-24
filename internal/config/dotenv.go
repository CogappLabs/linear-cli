package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadDotEnv walks up from the current directory looking for .env / .env.local
// and loads them into the process environment. Existing OS env vars win.
// Called at startup so LINEAR_API_KEY in a project .env is picked up.
func LoadDotEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		for _, name := range []string{".env.local", ".env"} {
			p := filepath.Join(dir, name)
			if _, err := os.Stat(p); err == nil {
				// godotenv.Load does not override already-set vars.
				_ = godotenv.Load(p)
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}
