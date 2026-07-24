package gitutil

import "testing"

func TestIdentifierFromBranch(t *testing.T) {
	cases := []struct {
		branch string
		want   string
	}{
		{"luke/abc-123-some-title", "ABC-123"},
		{"ENG-45", "ENG-45"},
		{"feature/eng-9-fix", "ENG-9"},
		{"lukew/luke-859-linear-cli-smoke-test", "LUKE-859"},
		{"main", ""},
		{"", ""},
		{"no-ticket-here", ""}, // "no-ticket" has no digits after the dash
		{"release/v2-0-1", "V2-0"},
	}
	for _, c := range cases {
		if got := identifierFromBranch(c.branch); got != c.want {
			t.Errorf("identifierFromBranch(%q) = %q, want %q", c.branch, got, c.want)
		}
	}
}
