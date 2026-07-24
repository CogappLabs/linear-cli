package config

import "testing"

func TestResolveAPIKeyPrefersEnv(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "env-key")
	if got := ResolveAPIKey(); got != "env-key" {
		t.Errorf("ResolveAPIKey() = %q, want env value", got)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	// Redirect the config dir to a temp location so the real config is untouched.
	t.Setenv("LINEAR_CLI_CONFIG_DIR", t.TempDir())
	t.Setenv("LINEAR_API_KEY", "") // ensure Load path is exercised, not env

	want := &Config{APIKey: "abc123", DefaultTeam: "ENG"}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.APIKey != want.APIKey || got.DefaultTeam != want.DefaultTeam {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, want)
	}
}
