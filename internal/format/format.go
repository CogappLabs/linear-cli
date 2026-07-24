package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	Bold   = color.New(color.Bold).SprintFunc()
	Gray   = color.New(color.FgHiBlack).SprintFunc()
	Red    = color.New(color.FgRed).SprintFunc()
	Green  = color.New(color.FgGreen).SprintFunc()
	Yellow = color.New(color.FgYellow).SprintFunc()
	Blue   = color.New(color.FgBlue).SprintFunc()
	Cyan   = color.New(color.FgCyan).SprintFunc()
)

// ColorState colours a workflow state by its Linear state type.
func ColorState(name, stateType string) string {
	switch stateType {
	case "completed":
		return Green(name)
	case "started":
		return Yellow(name)
	case "canceled":
		return Gray(name)
	case "unstarted":
		return Blue(name)
	default:
		return name
	}
}

// Priority renders Linear's 0-4 priority as a label.
func Priority(p int) string {
	switch p {
	case 1:
		return "Urgent"
	case 2:
		return "High"
	case 3:
		return "Medium"
	case 4:
		return "Low"
	default:
		return "None"
	}
}

// TimeAgo formats an ISO 8601 timestamp as a relative duration.
func TimeAgo(iso string) string {
	if iso == "" {
		return "never"
	}
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// PadRight pads s to width n, ignoring ANSI escapes when measuring.
func PadRight(s string, n int) string {
	visible := stripANSI(s)
	if len(visible) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(visible))
}

func stripANSI(s string) string {
	out := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			i = j + 1
			continue
		}
		out = append(out, s[i])
		i++
	}
	return string(out)
}
