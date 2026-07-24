package gitutil

import (
	"os/exec"
	"regexp"
	"strings"
)

// identifierRe matches a Linear issue identifier like ABC-123 anywhere in a
// branch name (Linear branch names look like "luke/abc-123-some-title").
var identifierRe = regexp.MustCompile(`(?i)([a-z][a-z0-9]*-\d+)`)

// CurrentBranch returns the current git branch, or "" if not in a repo.
func CurrentBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// IssueIDFromBranch extracts an ABC-123 style identifier from the current
// branch name, upper-cased. Returns "" if none found.
func IssueIDFromBranch() string {
	m := identifierRe.FindString(CurrentBranch())
	if m == "" {
		return ""
	}
	return strings.ToUpper(m)
}
